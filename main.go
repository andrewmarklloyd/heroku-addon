package main

import (
	"fmt"

	"github.com/andrewmarklloyd/heroku-addon/internal/pkg/config"
	"github.com/andrewmarklloyd/heroku-addon/internal/pkg/crypto"
	"github.com/andrewmarklloyd/heroku-addon/internal/pkg/heroku"
	"github.com/andrewmarklloyd/heroku-addon/internal/pkg/postgres"
	"github.com/andrewmarklloyd/heroku-addon/internal/pkg/web"
	"go.uber.org/zap"
)

var logger *zap.SugaredLogger

func main() {
	l, _ := zap.NewProduction()
	logger = l.Sugar().Named("heroku-addon")
	defer logger.Sync()

	cfg, err := config.BuildConfig()
	if err != nil {
		logger.Fatalln("error building config: %s", err.Error())
	}

	cryptoUtil, err := crypto.NewUtil(cfg.DBEncryptionKey)
	if err != nil {
		logger.Fatalln(fmt.Errorf("error creating crypto client: %s", err))
	}

	postgresClient, err := postgres.NewPostgresClient(cfg.PostgresURL)
	if err != nil {
		logger.Fatalln(fmt.Errorf("error creating postgres client: %s", err))
	}

	herokuClient := heroku.NewHerokuClient(cfg.Heroku.ClientSecret, cfg.Heroku.AddonUsername, cfg.Heroku.AddonPassword, cfg.Heroku.SSOSalt)

	webServer, err := web.NewWebServer(logger, cfg, cryptoUtil, postgresClient, herokuClient)
	if err != nil {
		logger.Fatalf("creating web server: %w", err)
	}

	err = webServer.HttpServer.ListenAndServe()
	if err != nil {
		logger.Fatalf("starting web server: %w", err)
	}
}
