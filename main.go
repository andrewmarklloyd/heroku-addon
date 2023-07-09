package main

import (
	"go.uber.org/zap"
)

var (
	logger *zap.SugaredLogger
)

func main() {
	l, _ := zap.NewProduction()
	logger = l.Sugar().Named("my-test-addon")
	defer logger.Sync()

	webServer, err := newWebServer()
	if err != nil {
		logger.Fatalf("creating web server: %w", err)
	}

	logger.Info("started web server")
	err = webServer.httpServer.ListenAndServe()
	if err != nil {
		logger.Fatalf("starting web server: %w", err)
	}
}
