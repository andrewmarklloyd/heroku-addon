package config

type Server struct {
	Port            string
	DBEncryptionKey string
	PostgresURL     string
	SessionSecret   SessionSecret
	Github          Github
	Heroku          Heroku
	Stripe          Stripe
}

type SessionSecret struct {
	HashKey       string
	EncryptionKey string
}

type Github struct {
	ClientID      string
	ClientSecret  string
	RedirectURL   string
	SessionSecret string
}

type Heroku struct {
	AddonUsername string
	AddonPassword string
	ClientSecret  string
	SSOSalt       string
}

type Stripe struct {
	Key                  string
	WebhookSigningSecret string
}
