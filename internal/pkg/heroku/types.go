package heroku

type PlanProvisionPayload struct {
	Plan       string     `json:"plan"`
	Region     string     `json:"region"`
	UUID       string     `json:"uuid"`
	OauthGrant OauthGrant `json:"oauth_grant"`
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	UserID       string `json:"user_id"`
	SessionNonce string `json:"session_nonce"`
}

type OauthResponse struct {
	Id           string `json:"id"`
	Message      string `json:"message"`
	AccessToken  string `json:"access_token"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
}

type OauthGrant struct {
	Code      string `json:"code"`
	ExpiresAt string `json:"expires_at"`
	Type      string `json:"type"`
}

type AppCollaborator struct {
	User AppUser `json:"user"`
	Role string  `json:"role"`
}

type AppUser struct {
	Email string `json:"email"`
}

type AddonInfo struct {
	App struct {
		Id   string `json:"id"`
		Name string `json:"name"`
	} `json:"app"`
}

type SSOUser struct {
	Email  string
	UserID string
	App    string
}
