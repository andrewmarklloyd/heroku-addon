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
	Name         string
	AccountType  AccountType
	AccessToken  string
	RefreshToken string
	StripeCustID string
}

type Instance struct {
	AccountID string `json:"accountID"`
	Id        string `json:"id"`
	Plan      string `json:"plan"`
	Name      string `json:"name"`
}

type PricingPlan struct {
	Name         string `json:"name"`
	PriceID      string `json:"priceID"`
	PriceDollars int    `json:"price"`
}

var PricingPlans = []PricingPlan{
	{
		Name:         "free",
		PriceID:      "price_1NpNZUGmaA1TfgH4vQUS0mw3",
		PriceDollars: 0,
	},
	{
		Name:         "staging",
		PriceID:      "price_1NpNZUGmaA1TfgH41kduGxJ8",
		PriceDollars: 10,
	},
	{
		Name:         "production",
		PriceID:      "price_1NpNZUGmaA1TfgH4yXZg6urh",
		PriceDollars: 35,
	},
}

func LookupPricingPlan(name string) PricingPlan {
	for _, plan := range PricingPlans {
		if plan.Name == name {
			return plan
		}
	}
	return PricingPlan{}
}
