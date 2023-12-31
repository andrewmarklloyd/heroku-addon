package web

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/andrewmarklloyd/heroku-addon/internal/pkg/account"
	"github.com/andrewmarklloyd/heroku-addon/internal/pkg/config"
	"github.com/andrewmarklloyd/heroku-addon/internal/pkg/crypto"
	"github.com/andrewmarklloyd/heroku-addon/internal/pkg/datadog"
	"github.com/andrewmarklloyd/heroku-addon/internal/pkg/heroku"
	"github.com/andrewmarklloyd/heroku-addon/internal/pkg/postgres"
	"github.com/andrewmarklloyd/heroku-addon/internal/pkg/provisioner"
	"github.com/andrewmarklloyd/heroku-addon/internal/pkg/spa"
	"github.com/dghubble/gologin"
	"github.com/dghubble/gologin/github"
	oauth2Login "github.com/dghubble/gologin/oauth2"
	"github.com/dghubble/sessions"
	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v75"
	"github.com/stripe/stripe-go/v75/customer"
	"golang.org/x/oauth2"
	githubOAuth2 "golang.org/x/oauth2/github"

	"go.uber.org/zap"

	gmux "github.com/gorilla/mux"
)

const (
	post   = "post"
	get    = "get"
	delete = "delete"
)

type ContextKey string

const ContextProvenanceKey ContextKey = "provenance"

type WebServer struct {
	HttpServer                 *http.Server
	cryptoUtil                 crypto.Util
	postgresClient             postgres.Client
	herokuClient               heroku.HerokuClient
	ddClient                   datadog.Client
	sessionStore               sessions.Store[string]
	logger                     *zap.SugaredLogger
	stripeKey                  string
	stripeWebhookSigningSecret string
	env                        string
}

func NewWebServer(logger *zap.SugaredLogger,
	cfg config.Server,
	cryptoUtil crypto.Util,
	postgresClient postgres.Client,
	herokuClient heroku.HerokuClient,
	ddClient datadog.Client,
	env string) (WebServer, error) {
	w := WebServer{
		cryptoUtil:                 cryptoUtil,
		postgresClient:             postgresClient,
		herokuClient:               herokuClient,
		ddClient:                   ddClient,
		logger:                     logger,
		stripeKey:                  cfg.Stripe.Key,
		stripeWebhookSigningSecret: cfg.Stripe.WebhookSigningSecret,
		env:                        env,
	}

	oauth2Config := &oauth2.Config{
		ClientID:     cfg.Github.ClientID,
		ClientSecret: cfg.Github.ClientSecret,
		RedirectURL:  cfg.Github.RedirectURL,
		Endpoint:     githubOAuth2.Endpoint,
		Scopes:       []string{"read:user", "user:email"},
	}

	// todo: make adding routes easier to see
	router := gmux.NewRouter().StrictSlash(true)

	// misc
	router.Handle("/health", http.HandlerFunc(healthHandler)).Methods(get)
	router.Handle("/login", http.HandlerFunc(w.tmpHandler)).Methods(get)

	// heroku
	router.Handle("/heroku/resources", w.requireHerokuAuth(http.HandlerFunc(w.provisionHerokuHandler))).Methods(post)
	router.Handle("/heroku/resources/{resource_uuid}", w.requireHerokuAuth(http.HandlerFunc(w.deprovisionHerokuHandler))).Methods(delete)

	store := sessions.NewCookieStore[string](
		sessions.DefaultCookieConfig,
		[]byte(cfg.SessionSecret.HashKey),
		[]byte(cfg.SessionSecret.EncryptionKey),
	)

	w.sessionStore = store
	router.Handle("/heroku/sso/login", http.HandlerFunc(w.herokuSSOHandler)).Methods(post)

	stateConfig := gologin.DefaultCookieConfig
	router.Handle("/github/login", github.StateHandler(stateConfig, github.LoginHandler(oauth2Config, nil)))
	router.Handle("/github/callback", github.StateHandler(stateConfig, github.CallbackHandler(oauth2Config, http.HandlerFunc(w.loginGithub), nil)))
	router.Handle("/logout", w.requireLogin(w.logout()))

	router.Handle("/api/user", w.requireLogin(http.HandlerFunc(w.getUser))).Methods(get)
	router.Handle("/api/pricing", http.HandlerFunc(w.getPricing)).Methods(get)
	router.Handle("/api/instances", w.requireLogin(http.HandlerFunc(w.getInstances))).Methods(get)
	router.Handle("/api/delete-instance", w.requireLogin(http.HandlerFunc(w.deleteInstance))).Methods(post)
	router.Handle("/api/create-payment-intent", w.requireLogin(http.HandlerFunc(w.newPaymentIntent))).Methods(post)
	router.Handle("/api/create-subscription", w.requireLogin(http.HandlerFunc(w.createSubscription))).Methods(post)
	router.Handle("/stripe-webhooks", http.HandlerFunc(w.handleStripeWebhook)).Methods(post)

	spa := spa.SpaHandler{
		StaticPath: "frontend/build",
		IndexPath:  "index.html",
	}

	router.PathPrefix("/").Handler(w.requireLogin(spa))

	addr := fmt.Sprintf("0.0.0.0:%s", cfg.Port)
	logger.Infof("starting web server on address %s", addr)
	server := &http.Server{
		Handler: router,
		Addr:    addr,
	}

	w.HttpServer = server
	return w, nil
}

func (s WebServer) herokuSSOHandler(w http.ResponseWriter, req *http.Request) {
	s.ddClient.Publish(req.Context(), datadog.CustomMetric{
		MetricName:  datadog.MetricNameLogin,
		MetricValue: 1,
		Tags: map[string]string{
			"login_source": string(datadog.MetricTagHeroku),
		},
	})

	ssoUser, err := s.herokuClient.ValidateSSO(req)
	if err != nil {
		s.logger.Errorf("validating heroku sso: %s", err)
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(`<!DOCTYPE html><html><h1>forbidden</h1></html>`))
		return
	}

	a, err := s.postgresClient.GetAccountFromEmail(s.cryptoUtil, ssoUser.Email, string(account.AccountTypeHeroku))
	if err != nil {
		s.logger.Errorf("getting heroku account from email: %s", err)
		http.Redirect(w, req, "/login", http.StatusFound)
		return
	}

	session := s.sessionStore.New("heroku-addon")
	session.Set("user-email", ssoUser.Email)
	session.Set("user-id", a.UUID)
	session.Set("user-name", a.Name)
	session.Set("provenance", "heroku")
	if err := session.Save(w); err != nil {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(`<!DOCTYPE html><html><h1>forbidden</h1></html>`))
		return
	}

	http.Redirect(w, req, "/", http.StatusFound)
}

func (s WebServer) tmpHandler(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`<!DOCTYPE html>
	<html lang="en">
	  <head>
		<title>Nothing</title>
	  </head>
	  <style>
	  .button {
		background-color: #555555; /* Black */
		border: none;
		color: white;
		padding: 15px 32px;
		text-align: center;
		text-decoration: none;
		display: inline-block;
		font-size: 16px;
	  }
	  </style>
	  <body>
		<div id="root">
		  <h1>Login</h1>
		</div>
		<div>
			<button class="button" onclick="window.location.href='/github/login';">
			Login with Github
			</button>
		</div>
		<br></br>
		<div>
			<button class="button" onclick="window.location.href='https://elements.heroku.com/addons/alloyd-poc';">
			Provision Addon in Heroku
			</button>
		</div>
		<h2 id="login-failed-reason"></h2>
	  </body>
	  <script>
	  	const urlParams = new URLSearchParams(window.location.search)
		const reason = urlParams.get('reason')
	  	if (reason) {
			const el = document.getElementById("login-failed-reason")
			el.style.display = "block";
			el.innerHTML = "LOGIN FAILED<br/>" + reason;
		}
	  </script>
	</html>
	`))
}

func (s WebServer) loginGithub(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	s.ddClient.Publish(req.Context(), datadog.CustomMetric{
		MetricName:  datadog.MetricNameLogin,
		MetricValue: 1,
		Tags: map[string]string{
			"login_source": string(datadog.MetricTagGithub),
		},
	})

	user, err := github.UserFromContext(ctx)
	if err != nil {
		s.errorLogAndRedirect(w, req, fmt.Sprintf("getting user from context: %s", err), "user authentication error")
		return
	}

	token, err := oauth2Login.TokenFromContext(ctx)
	if err != nil {
		s.errorLogAndRedirect(w, req, fmt.Sprintf("getting github token from context: %s", err.Error()), "Could not get token from Github")
		return
	}
	if token == nil {
		s.errorLogAndRedirect(w, req, "github token is nil, cannot login", "Could not get token from Github")
		return
	}

	email, err := getGithubUserEmail(ctx, *token)
	if err != nil {
		s.errorLogAndRedirect(w, req, fmt.Sprintf("getting github user email: %s", err.Error()), "Could not user email from Github")
		return
	}

	if !strings.Contains(os.Getenv("AUTHORIZED_USERS"), email) {
		s.errorLogAndRedirect(w, req, fmt.Sprintf("non authorized user attempted login: %s", email), "User is not authorized")
		return
	}

	userName := user.GetName()
	if userName == "" {
		userName = "Github User"
	}

	a, err := s.postgresClient.GetAccountFromEmail(s.cryptoUtil, email, string(account.AccountTypeGithub))
	if err != nil {
		var noAcctErr *postgres.AccountNotFound
		if errors.As(err, &noAcctErr) {
			stripe.Key = s.stripeKey
			params := &stripe.CustomerParams{
				Name:  stripe.String(userName),
				Email: stripe.String(email),
				Metadata: map[string]string{
					"env": s.env,
				},
			}
			cust, err := customer.New(params)
			if err != nil {
				s.logger.Errorf("creating new stripe customer: %s", err)
				http.Redirect(w, req, "/login", http.StatusFound)
				return
			}

			id := uuid.New().String()
			a = account.Account{
				UUID:         id,
				Email:        email,
				Name:         userName,
				AccountType:  account.AccountTypeGithub,
				AccessToken:  "",
				RefreshToken: "",
				StripeCustID: cust.ID,
			}

			err = s.postgresClient.CreateOrUpdateAccount(s.cryptoUtil, a)
			if err != nil {
				s.logger.Errorf("creating new account: %s", err)
				http.Redirect(w, req, "/login", http.StatusFound)
				return
			}
		} else {
			s.logger.Errorf("getting account from email: %s", err)
			http.Redirect(w, req, "/login", http.StatusFound)
			return
		}
	}

	session := s.sessionStore.New("heroku-addon")
	session.Set("user-email", email)
	session.Set("user-id", a.UUID)
	session.Set("user-name", a.Name)
	session.Set("stripe-id", a.StripeCustID)
	session.Set("provenance", "github")
	if err := session.Save(w); err != nil {
		s.logger.Errorf("saving session: %s", err)
		http.Redirect(w, req, "/login", http.StatusFound)
		return
	}

	http.Redirect(w, req, "/", http.StatusFound)
}

func (s WebServer) logout() http.Handler {
	fn := func(w http.ResponseWriter, req *http.Request) {
		s.sessionStore.Destroy(w, "heroku-addon")
		http.Redirect(w, req, "/login", http.StatusFound)
	}
	return http.HandlerFunc(fn)
}

func (s WebServer) requireLogin(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		session, err := s.sessionStore.Get(r, "heroku-addon")

		if err != nil {
			// s.logger.Errorf("could not get session: %s", err)
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		provenance, ok := session.GetOk("provenance")
		if !ok {
			s.logger.Errorf("could not get provenance from context")
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		ctx := r.Context()
		req := r.WithContext(context.WithValue(ctx, ContextProvenanceKey, provenance))
		*r = *req

		_, present := session.GetOk("user-email")
		if !present {
			s.logger.Errorf("could not get user-email: %s", err)
			http.Redirect(w, req, "/login", http.StatusFound)
			return
		}

		next.ServeHTTP(w, req)

	}
	return http.HandlerFunc(fn)
}

func (s WebServer) provisionHerokuHandler(w http.ResponseWriter, req *http.Request) {
	s.logger.Infof("got request for new provisioning")
	s.ddClient.Publish(req.Context(), datadog.CustomMetric{
		MetricName:  datadog.MetricNameProvision,
		MetricValue: 1,
		Tags: map[string]string{
			"type": "heroku",
		},
	})

	var payload heroku.PlanProvisionPayload
	err := json.NewDecoder(req.Body).Decode(&payload)
	if err != nil {
		s.logger.Errorf("Error parsing payload: %s", err)
		http.Error(w, `{"error":"Error parsing request","status":"failed"}`, http.StatusBadRequest)
		return
	}

	s.logger.Infof("starting provision process for %s", payload.UUID)

	oauthResp, err := s.herokuClient.ExchangeToken(payload.OauthGrant.Code)
	if err != nil {
		s.logger.Errorf("error exchanging token: %s", err)
		emptyOauthResp := heroku.OauthResponse{}
		if oauthResp != emptyOauthResp {
			s.logger.Errorf("token response: %s", oauthResp)
		}
		http.Error(w, `{"error":"error exchanging token","status":"failed"}`, http.StatusBadRequest)
		return
	}

	addonInfo, err := s.herokuClient.GetAppAddonInfo(oauthResp.AccessToken)
	if err != nil {
		s.logger.Errorf("error getting app id: %s", err)
		http.Error(w, `{"error":"error provisioning","status":"failed"}`, http.StatusInternalServerError)
		return
	}

	ownerEmail, err := heroku.GetOwnerEmail(oauthResp.AccessToken, addonInfo.App.Id)
	if err != nil {
		s.logger.Errorf("error getting owner email: %s", err)
		http.Error(w, `{"error":"error provisioning","status":"failed"}`, http.StatusInternalServerError)
		return
	}

	acct := account.Account{
		UUID:         payload.UUID,
		Email:        ownerEmail,
		Name:         addonInfo.App.Name,
		AccountType:  account.AccountTypeHeroku,
		AccessToken:  oauthResp.AccessToken,
		RefreshToken: oauthResp.RefreshToken,
		StripeCustID: "", // payment handled by Heroku, not required
	}
	err = s.postgresClient.CreateOrUpdateAccount(s.cryptoUtil, acct)
	if err != nil {
		s.logger.Errorf("error creating account: %s", err)
		http.Error(w, `{"error":"error provisioning","status":"failed"}`, http.StatusInternalServerError)
		return
	}

	idAndName := uuid.New().String()
	a := account.Instance{
		AccountID: payload.UUID,
		Id:        idAndName,
		Plan:      payload.Plan,
		Name:      idAndName,
	}

	err = s.postgresClient.CreateOrUpdateInstance(a)
	if err != nil {
		s.logger.Errorf("error creating instance: %s", err)
		http.Error(w, `{"error":"saving instance to database","status":"failed"}`, http.StatusInternalServerError)
		return
	}

	err = provisioner.ProvisionResource()
	if err != nil {
		s.logger.Errorf("error provisioning resource: %s", err)
		http.Error(w, `{"error":"error provisioning","status":"failed"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	// w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte(fmt.Sprintf(`{"id":"%s","message":"Your add-on is provisioned!","config": { "TESTING": "hello" }
	}`, payload.UUID)))
}

func (s WebServer) deprovisionHerokuHandler(w http.ResponseWriter, req *http.Request) {
	s.logger.Infof("got request to delete addon")
	s.ddClient.Publish(req.Context(), datadog.CustomMetric{
		MetricName:  datadog.MetricNameDeprovision,
		MetricValue: 1,
		Tags: map[string]string{
			"type": "heroku",
		},
	})

	vars := gmux.Vars(req)

	s.logger.Infof("deleting heroku addon instance", "resource_uuid", vars["resource_uuid"])

	// delete instance
	// keep account in case return customer?

	w.WriteHeader(http.StatusNoContent)
	// TODO: add correct id
	w.Write([]byte(`{"id":"hello","message":"Your add-on has been deleted."}`))

}

func (s *WebServer) requireHerokuAuth(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, req *http.Request) {
		if !s.herokuClient.ValidateBasicAuth(req) {
			return
		}

		next.ServeHTTP(w, req)
	}
	return http.HandlerFunc(fn)
}

func healthHandler(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, `ok`)
}

func (s WebServer) getPricing(w http.ResponseWriter, req *http.Request) {
	plans, err := json.Marshal(account.PricingPlans)
	if err != nil {
		s.logger.Errorf("marshalling pricing plans to json: %s", err)
		http.Error(w, "could not get pricing", http.StatusInternalServerError)
		return
	}
	fmt.Fprint(w, string(plans))
}
