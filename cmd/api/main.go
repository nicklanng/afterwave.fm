package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	"github.com/kelseyhightower/envconfig"
	apphttp "github.com/sopatech/afterwave.fm/internal/http"
	"github.com/sopatech/afterwave.fm/internal/infra"
	"github.com/sopatech/afterwave.fm/internal/users"
)

type Config struct {
	Addr        string `envconfig:"ADDR" default:":8080"`
	AWSRegion   string `envconfig:"AWS_REGION" default:"us-east-1"`
	DynamoTable string `envconfig:"DYNAMO_TABLE" required:"true"`
	JWTSecret   string `envconfig:"JWT_SECRET" required:"true"`
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		logger.Error("config", "err", err)
		os.Exit(1)
	}

	db, err := infra.NewDynamo(context.Background(), cfg.AWSRegion)
	if err != nil {
		logger.Error("dynamo init", "err", err)
		os.Exit(1)
	}

	userSvc := users.NewService(db, cfg.DynamoTable, []byte(cfg.JWTSecret))
	r := apphttp.NewRouter(logger, users.NewHandler(userSvc), cfg.JWTSecret)

	logger.Info("listening", "addr", cfg.Addr)
	if err := http.ListenAndServe(cfg.Addr, r); err != nil {
		logger.Error("server", "err", err)
		os.Exit(1)
	}
}
