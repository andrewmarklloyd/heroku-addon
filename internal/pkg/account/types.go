package account

type AccountType string

const (
	AccountTypeGithub AccountType = "github"
	AccountTypeHeroku AccountType = "heroku"
)

type PlanType string

const (
	PlanTypeFree       PlanType = "free"
	PlanTypeStaging    PlanType = "staging"
	PlanTypeProduction PlanType = "production"
)

type Account struct {
	UUID         string
	Email        string
	AccountType  AccountType
	PlanType     PlanType
	AccessToken  string
	RefreshToken string
}
