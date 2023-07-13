package config

type Server struct {
	Port            string
	DBEncryptionKey string
	PostgresURL     string
	SessionSecret   SessionSecret
	Github          Github
	Heroku          Heroku
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
