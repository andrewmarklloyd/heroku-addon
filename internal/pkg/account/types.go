package account

type AccountType string

const (
	AccountTypeGithub AccountType = "github"
	AccountTypeHeroku AccountType = "heroku"
)

type PlanType string

const (
	PlanTypeTest       PlanType = "test"
	PlanTypeFree       PlanType = "free"
	PlanTypeStaging    PlanType = "staging"
	PlanTypeProduction PlanType = "production"
)

type Account struct {
	UUID         string
	Email        string
	AccountType  AccountType
	AccessToken  string
	RefreshToken string
}

type Instance struct {
	AccountID string `json:"accountID"`
	Id        string `json:"id"`
	Plan      string `json:"plan"`
	Name      string `json:"name"`
}
