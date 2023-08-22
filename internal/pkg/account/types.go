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
	Name    string `json:"name"`
	PriceID string `json:"priceID"`
	Price   int    `json:"price"`
}

var PricingPlans = []PricingPlan{
	{
		Name:    "free",
		PriceID: "price_1NhPmEGmaA1TfgH4aPVJx0LZ",
		Price:   0,
	},
	{
		Name:    "staging",
		PriceID: "price_1NhPndGmaA1TfgH4imxsTGqA",
		Price:   10,
	},
	{
		Name:    "production",
		PriceID: "price_1NgKTiGmaA1TfgH4Fr6itojs",
		Price:   35,
	},
}
