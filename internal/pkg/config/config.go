package config

import (
	"errors"
	"fmt"
	"os"
)

func BuildConfig() (Server, error) {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	var err error

	encKey := os.Getenv("ENCRYPTION_KEY")
	if encKey == "" {
		err = errors.Join(err, fmt.Errorf("ENCRYPTION_KEY env var is not set"))
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		err = errors.Join(err, fmt.Errorf("DATABASE_URL env var is not set"))
	}

	sessHashKey := os.Getenv("SESSION_SECRET_HASH_KEY")
	if sessHashKey == "" {
		err = errors.Join(err, fmt.Errorf("SESSION_SECRET_HASH_KEY env var is not set"))
	}

	sessEncKey := os.Getenv("SESSION_SECRET_ENCRYPTION_KEY")
	if sessEncKey == "" {
		err = errors.Join(err, fmt.Errorf("SESSION_SECRET_HASH_KEY env var is not set"))
	}

	herokuAddonUsername := os.Getenv("HEROKU_ADDON_USERNAME")
	if herokuAddonUsername == "" {
		err = errors.Join(err, fmt.Errorf("HEROKU_ADDON_USERNAME env var is not set"))
	}

	herokuAddonPassword := os.Getenv("HEROKU_ADDON_PASSWORD")
	if herokuAddonPassword == "" {
		err = errors.Join(err, fmt.Errorf("HEROKU_ADDON_PASSWORD env var is not set"))
	}

	herokuClientSecret := os.Getenv("HEROKU_CLIENT_SECRET")
	if herokuClientSecret == "" {
		err = errors.Join(err, fmt.Errorf("HEROKU_CLIENT_SECRET env var is not set"))
	}

	githubClientID := os.Getenv("GITHUB_CLIENT_ID")
	if githubClientID == "" {
		err = errors.Join(err, fmt.Errorf("GITHUB_CLIENT_ID env var is not set"))
	}

	githubClientSecret := os.Getenv("GITHUB_CLIENT_SECRET")
	if githubClientSecret == "" {
		err = errors.Join(err, fmt.Errorf("GITHUB_CLIENT_SECRET env var is not set"))
	}

	githubRedirectURL := os.Getenv("GITHUB_REDIRECT_URI")
	if githubRedirectURL == "" {
		err = errors.Join(err, fmt.Errorf("GITHUB_REDIRECT_URI env var is not set"))
	}

	if err != nil {
		return Server{}, err
	}

	return Server{
		Port:            port,
		DBEncryptionKey: encKey,
		PostgresURL:     dbURL,
		SessionSecret: SessionSecret{
			HashKey:       sessHashKey,
			EncryptionKey: sessEncKey,
		},
		Heroku: Heroku{
			AddonUsername: herokuAddonUsername,
			AddonPassword: herokuAddonPassword,
			ClientSecret:  herokuClientSecret,
		},
		Github: Github{
			ClientID:     githubClientID,
			ClientSecret: githubClientSecret,
			RedirectURL:  githubRedirectURL,
		},
	}, nil
}
