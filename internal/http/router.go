package http

import (
	"crypto/rsa"
	"log/slog"
	"net/http"

	authmw "github.com/sopatech/afterwave.fm/internal/auth"
	"github.com/sopatech/afterwave.fm/internal/artists"
	"github.com/sopatech/afterwave.fm/internal/feed"
	"github.com/sopatech/afterwave.fm/internal/follows"
	"github.com/sopatech/afterwave.fm/internal/users"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func chain(h http.Handler, mws ...func(http.Handler) http.Handler) http.Handler {
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
}

func NewRouter(logger *slog.Logger, userH *users.Handler, authH *authmw.Handler, artistH *artists.Handler, followH *follows.Handler, feedH *feed.Handler, jwtPublicKey *rsa.PublicKey) http.Handler {
	mux := http.NewServeMux()

	wrap := func(h http.Handler) http.Handler {
		return chain(h,
			Recoverer(logger),
			RealIP,
			RequestLogger(logger),
		)
	}

	auth := authmw.Authenticate(jwtPublicKey)

	// v1 API
	v1 := http.NewServeMux()

	// Public auth (Authorization Code + PKCE; no client secret)
	v1.Handle("POST /auth/signup", wrap(http.HandlerFunc(userH.Signup)))
	v1.Handle("POST /auth/login", wrap(http.HandlerFunc(userH.Login)))
	v1.Handle("POST /auth/token", wrap(http.HandlerFunc(authH.Token)))
	v1.Handle("POST /auth/refresh", wrap(http.HandlerFunc(authH.Refresh)))
	v1.Handle("POST /auth/logout", wrap(auth(http.HandlerFunc(authH.Logout))))

	// Protected
	v1.Handle("GET /users/me", wrap(auth(http.HandlerFunc(userH.Me))))
	v1.Handle("DELETE /account", wrap(auth(http.HandlerFunc(userH.DeleteAccount))))

	// Following and my feed
	v1.Handle("POST /users/me/following/{handle}", wrap(auth(http.HandlerFunc(followH.Follow))))
	v1.Handle("DELETE /users/me/following/{handle}", wrap(auth(http.HandlerFunc(followH.Unfollow))))
	v1.Handle("GET /users/me/following", wrap(auth(http.HandlerFunc(followH.ListFollowing))))
	v1.Handle("GET /feed", wrap(auth(http.HandlerFunc(feedH.MyFeed))))

	// Artists: protected create/list-mine; public get-by-handle; protected update/delete (owner only)
	v1.Handle("POST /artists", wrap(auth(http.HandlerFunc(artistH.Create))))
	v1.Handle("GET /artists/me", wrap(auth(http.HandlerFunc(artistH.ListMine))))
	v1.Handle("GET /artists/{handle}", wrap(http.HandlerFunc(artistH.GetByHandle)))
	v1.Handle("PATCH /artists/{handle}", wrap(auth(http.HandlerFunc(artistH.Update))))
	v1.Handle("DELETE /artists/{handle}", wrap(auth(http.HandlerFunc(artistH.Delete))))

	// Feed (posts): public list/get; protected create/update/delete (owner only)
	v1.Handle("POST /artists/{handle}/posts", wrap(auth(http.HandlerFunc(feedH.CreatePost))))
	v1.Handle("GET /artists/{handle}/posts", wrap(http.HandlerFunc(feedH.ListPosts)))
	v1.Handle("GET /artists/{handle}/posts/{postId}", wrap(http.HandlerFunc(feedH.GetPost)))
	v1.Handle("PATCH /artists/{handle}/posts/{postId}", wrap(auth(http.HandlerFunc(feedH.UpdatePost))))
	v1.Handle("DELETE /artists/{handle}/posts/{postId}", wrap(auth(http.HandlerFunc(feedH.DeletePost))))

	mux.Handle("/v1/", http.StripPrefix("/v1", v1))

	return otelhttp.NewHandler(mux, "http.server")
}
