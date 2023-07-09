package main

import (
	"fmt"
	"os"

	"github.com/andrewmarklloyd/heroku-addon/internal/pkg/crypto"
	"github.com/andrewmarklloyd/heroku-addon/internal/pkg/heroku"
	"github.com/andrewmarklloyd/heroku-addon/internal/pkg/postgres"
	"github.com/andrewmarklloyd/heroku-addon/internal/pkg/web"
	"go.uber.org/zap"
)

var logger *zap.SugaredLogger

type serverConfig struct {
	encryptionKey       string
	dbURL               string
	herokuAddonUsername string
	herokuAddonPassword string
	herokuClientSecret  string
}

func main() {
	l, _ := zap.NewProduction()
	logger = l.Sugar().Named("heroku-addon")
	defer logger.Sync()

	cfg := buildConfig()

	cryptoUtil, err := crypto.NewUtil(cfg.encryptionKey)
	if err != nil {
		logger.Fatalln(fmt.Errorf("error creating crypto client: %s", err))
	}

	postgresClient, err := postgres.NewPostgresClient(cfg.dbURL)
	if err != nil {
		logger.Fatalln(fmt.Errorf("error creating postgres client: %s", err))
	}

	herokuClient := heroku.NewHerokuClient(cfg.herokuClientSecret, cfg.herokuAddonUsername, cfg.herokuAddonPassword)

	webServer, err := web.NewWebServer(cryptoUtil, postgresClient, herokuClient)
	if err != nil {
		logger.Fatalf("creating web server: %w", err)
	}

	logger.Info("started web server")
	err = webServer.HttpServer.ListenAndServe()
	if err != nil {
		logger.Fatalf("starting web server: %w", err)
	}
}

func buildConfig() serverConfig {
	encKey := os.Getenv("ENCRYPTION_KEY")
	if encKey == "" {
		logger.Fatalln("ENCRYPTION_KEY env var is not set")
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		logger.Fatalln("DATABASE_URL env var is not set")
	}

	herokuAddonUsername := os.Getenv("HEROKU_ADDON_USERNAME")
	if herokuAddonUsername == "" {
		logger.Fatalln("HEROKU_ADDON_USERNAME env var is not set")
	}

	herokuAddonPassword := os.Getenv("HEROKU_ADDON_PASSWORD")
	if herokuAddonPassword == "" {
		logger.Fatalln("HEROKU_ADDON_PASSWORD env var is not set")
	}

	herokuClientSecret := os.Getenv("HEROKU_CLIENT_SECRET")
	if herokuClientSecret == "" {
		logger.Fatalln("HEROKU_CLIENT_SECRET env var is not set")
	}

	return serverConfig{
		encryptionKey:       encKey,
		dbURL:               dbURL,
		herokuAddonUsername: herokuAddonUsername,
		herokuAddonPassword: herokuAddonPassword,
		herokuClientSecret:  herokuClientSecret,
	}
}
