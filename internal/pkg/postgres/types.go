package postgres

type Account struct {
	UUID         string
	AccessToken  string
	RefreshToken string
}
