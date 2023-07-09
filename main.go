package main

import (
	"fmt"
	"os"

	"github.com/andrewmarklloyd/heroku-addon/internal/pkg/crypto"
	"github.com/andrewmarklloyd/heroku-addon/internal/pkg/postgres"
	"github.com/andrewmarklloyd/heroku-addon/internal/pkg/web"
	"go.uber.org/zap"
)

func main() {
	l, _ := zap.NewProduction()
	logger := l.Sugar().Named("my-test-addon")
	defer logger.Sync()

	cryptoUtil, err := crypto.NewUtil(os.Getenv("ENCRYPTION_KEY"))
	if err != nil {
		logger.Fatalln(fmt.Errorf("error creating crypto client: %s", err))
	}

	postgresClient, err := postgres.NewPostgresClient(os.Getenv("DATABASE_URL"))
	if err != nil {
		logger.Fatalln(fmt.Errorf("error creating postgres client: %s", err))
	}

	webServer, err := web.NewWebServer(cryptoUtil, postgresClient)
	if err != nil {
		logger.Fatalf("creating web server: %w", err)
	}

	logger.Info("started web server")
	err = webServer.HttpServer.ListenAndServe()
	if err != nil {
		logger.Fatalf("starting web server: %w", err)
	}
}
