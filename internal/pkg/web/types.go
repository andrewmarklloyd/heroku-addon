package web

type GithubAuth struct {
	SessionSecret string
}

type UserInfo struct {
	UserID     string `json:"userID"`
	Email      string `json:"email"`
	Name       string `json:"name"`
	Provenance string `json:"provenance"`
	StripeID   string `json:"stripeID"`
}
