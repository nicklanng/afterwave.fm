package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/kelseyhightower/envconfig"
	"github.com/sopatech/afterwave.fm/internal/artists"
	"github.com/sopatech/afterwave.fm/internal/auth"
	"github.com/sopatech/afterwave.fm/internal/cognito"
	"github.com/sopatech/afterwave.fm/internal/config"
	"github.com/sopatech/afterwave.fm/internal/feed"
	"github.com/sopatech/afterwave.fm/internal/follows"
	apphttp "github.com/sopatech/afterwave.fm/internal/http"
	"github.com/sopatech/afterwave.fm/internal/infra"
	"github.com/sopatech/afterwave.fm/internal/metrics"
	"github.com/sopatech/afterwave.fm/internal/search"
	"github.com/sopatech/afterwave.fm/internal/users"

	"github.com/prometheus/client_golang/prometheus"
)

// Config holds process configuration from environment (envconfig).
type Config struct {
	Addr                string `envconfig:"ADDR" default:":8080"`
	AWSRegion           string `envconfig:"AWS_REGION" default:"us-east-1"`
	DynamoTable         string `envconfig:"DYNAMO_TABLE" default:"afterwave"`
	DynamoEndpoint      string `envconfig:"DYNAMODB_ENDPOINT"`            // optional, e.g. http://localhost:8001 for DynamoDB Local
	OpenSearchEndpoint  string `envconfig:"OPENSEARCH_ENDPOINT" required:"true"` // e.g. http://localhost:9200 for feed index and my feed
	OpenSearchFeedIndex string `envconfig:"OPENSEARCH_FEED_INDEX" default:"afterwave-feed"`
	JWTPrivateKeyPath   string `envconfig:"JWT_PRIVATE_KEY_PATH" required:"true"`
	JWTPublicKeyPath    string `envconfig:"JWT_PUBLIC_KEY_PATH" required:"true"`
	CookieSecure        bool   `envconfig:"COOKIE_SECURE" default:"true"`
	CognitoUserPoolID   string `envconfig:"COGNITO_USER_POOL_ID" required:"true"`
	CognitoClientID     string `envconfig:"COGNITO_CLIENT_ID" required:"true"`
	CognitoClientSecret string `envconfig:"COGNITO_CLIENT_SECRET"`                // optional; required for confidential app client token exchange
	CognitoHostedDomain string `envconfig:"COGNITO_HOSTED_UI_DOMAIN"`             // e.g. https://<domain>.auth.<region>.amazoncognito.com
	CognitoCallbackURL  string `envconfig:"COGNITO_CALLBACK_URL"`                 // e.g. https://api.afterwave.fm/v1/auth/callback
	FrontendRedirectURI string `envconfig:"FRONTEND_REDIRECT_URI"`                // e.g. https://app.afterwave.fm/auth/callback
	OAuthStateSecret    string `envconfig:"OAUTH_STATE_SECRET"`                   // optional; if set, federated flow validates CSRF state cookie
}

func main() {
	// --- Logger and config ---
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		logger.Error("config", "err", err)
		os.Exit(1)
	}
	config.LogConfigVars(logger, &cfg)

	// --- DynamoDB ---
	db, err := infra.NewDynamo(context.Background(), cfg.AWSRegion, cfg.DynamoEndpoint)
	if err != nil {
		logger.Error("dynamo init", "err", err)
		os.Exit(1)
	}

	// --- Auth: store, clients, JWT keys, service, handler ---
	authStore := auth.NewStore(db, cfg.DynamoTable)
	// Public clients (no secret): web 15min/7d, native 30d/90d
	const (
		webSessionSec, webRefreshSec        = 15 * 60, 7 * 24 * 3600
		nativeSessionSec, nativeRefreshSec  = 30 * 24 * 3600, 90 * 24 * 3600
	)
	authClients := []auth.ClientCredential{
		{ID: "web", SessionTTLSeconds: webSessionSec, RefreshTTLSeconds: webRefreshSec},
		{ID: "ios", SessionTTLSeconds: nativeSessionSec, RefreshTTLSeconds: nativeRefreshSec},
		{ID: "android", SessionTTLSeconds: nativeSessionSec, RefreshTTLSeconds: nativeRefreshSec},
		{ID: "desktop", SessionTTLSeconds: nativeSessionSec, RefreshTTLSeconds: nativeRefreshSec},
	}
	if err := authStore.EnsureAuthClients(context.Background(), authClients); err != nil {
		logger.Error("ensure auth clients", "err", err)
		os.Exit(1)
	}
	jwtPrivateKey, err := auth.LoadRSAPrivateKey(cfg.JWTPrivateKeyPath)
	if err != nil {
		logger.Error("load JWT private key", "err", err)
		os.Exit(1)
	}
	jwtPublicKey, err := auth.LoadRSAPublicKey(cfg.JWTPublicKeyPath)
	if err != nil {
		logger.Error("load JWT public key", "err", err)
		os.Exit(1)
	}
	// MAU: conditional put to DynamoDB; increment Prometheus counter only when row is new this month
	mauStore := metrics.NewMAUStore(db, cfg.DynamoTable)
	mauRecorder, err := metrics.NewMAURecorder(mauStore, prometheus.DefaultRegisterer)
	if err != nil {
		logger.Error("register MAU counter", "err", err)
		os.Exit(1)
	}
	authService := auth.NewService(authStore, jwtPrivateKey, mauRecorder)
	cookieCfg := auth.CookieConfig{Secure: cfg.CookieSecure}
	authHandler := auth.NewHandler(authService, cookieCfg)

	// --- Users: store, service, handler ---
	usersStore := users.NewStore(db, cfg.DynamoTable)
	cognitoClient, err := cognito.NewAWSClient(context.Background(), cfg.AWSRegion, cfg.CognitoUserPoolID, cfg.CognitoClientID)
	if err != nil {
		logger.Error("cognito init", "err", err)
		os.Exit(1)
	}
	usersService := users.NewService(usersStore, cognitoClient)
	usersHandler := users.NewHandler(usersService, authService, cookieCfg, cfg.CognitoHostedDomain, cfg.AWSRegion, cfg.CognitoUserPoolID, cfg.CognitoClientID, cfg.CognitoClientSecret, cfg.CognitoCallbackURL, cfg.FrontendRedirectURI, cfg.OAuthStateSecret)

	// --- Artists: store, service, handler ---
	artistsStore := artists.NewStore(db, cfg.DynamoTable)
	artistsMemberStore := artists.NewMemberStore(db, cfg.DynamoTable)
	artistsService := artists.NewService(artistsStore, artistsMemberStore)
	artistsHandler := artists.NewHandler(artistsService)

	// --- Follows: store, service, handler ---
	followsStore := follows.NewStore(db, cfg.DynamoTable)
	followsService := follows.NewService(followsStore, artistsService)
	followsHandler := follows.NewHandler(followsService)

	// --- Feed: store, OpenSearch index, service, handler ---
	feedStore := feed.NewStore(db, cfg.DynamoTable)
	osClient := infra.NewOpenSearch(cfg.OpenSearchEndpoint, nil)
	feedIndex := search.NewFeedIndex(osClient, cfg.OpenSearchFeedIndex)
	if err := feedIndex.EnsureIndex(context.Background()); err != nil {
		logger.Error("opensearch ensure feed index", "err", err)
		os.Exit(1)
	}
	feedService := feed.NewServiceWithSearch(feedStore, artistsService, artistsService, feedIndex, followsService, feedIndex)
	feedHandler := feed.NewHandler(feedService)

	// --- Router and HTTP server ---
	r := apphttp.NewRouter(logger, usersHandler, authHandler, artistsHandler, followsHandler, feedHandler, metrics.Handler(), jwtPublicKey)

	srv := &http.Server{
		Addr:         cfg.Addr,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	logger.Info("listening", "addr", cfg.Addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error("server", "err", err)
		os.Exit(1)
	}
}
