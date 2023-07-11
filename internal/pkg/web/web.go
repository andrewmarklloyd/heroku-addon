package web

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/andrewmarklloyd/heroku-addon/internal/pkg/config"
	"github.com/andrewmarklloyd/heroku-addon/internal/pkg/crypto"
	"github.com/andrewmarklloyd/heroku-addon/internal/pkg/heroku"
	"github.com/andrewmarklloyd/heroku-addon/internal/pkg/postgres"
	"github.com/andrewmarklloyd/heroku-addon/internal/pkg/provisioner"
	"github.com/andrewmarklloyd/heroku-addon/internal/pkg/spa"
	"github.com/dghubble/gologin"
	"github.com/dghubble/gologin/github"
	"github.com/dghubble/sessions"
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

type WebServer struct {
	HttpServer     *http.Server
	cryptoUtil     crypto.Util
	postgresClient postgres.Client
	herokuClient   heroku.HerokuClient
	sessions       sessions.Store[string]
	logger         *zap.SugaredLogger
}

func NewWebServer(logger *zap.SugaredLogger,
	cfg config.Server,
	cryptoUtil crypto.Util,
	postgresClient postgres.Client,
	herokuClient heroku.HerokuClient) (WebServer, error) {
	w := WebServer{
		cryptoUtil:     cryptoUtil,
		postgresClient: postgresClient,
		herokuClient:   herokuClient,
		logger:         logger,
	}

	oauth2Config := &oauth2.Config{
		ClientID:     cfg.Github.ClientID,
		ClientSecret: cfg.Github.ClientSecret,
		RedirectURL:  cfg.Github.RedirectURL,
		Endpoint:     githubOAuth2.Endpoint,
		Scopes:       []string{"profile", "email"},
	}

	router := gmux.NewRouter().StrictSlash(true)

	router.Handle("/health", http.HandlerFunc(healthHandler)).Methods(get)
	router.Handle("/welcome", http.HandlerFunc(w.tmpHandler)).Methods(get)

	router.Handle("/heroku/resources", w.requireHerokuAuth(http.HandlerFunc(w.provisionHandler))).Methods(post)
	router.Handle("/heroku/resources/{resource_uuid}", w.requireHerokuAuth(http.HandlerFunc(w.deprovisionHandler))).Methods(delete)

	store := sessions.NewCookieStore[string](
		sessions.DefaultCookieConfig,
		[]byte(cfg.SessionSecret.HashKey),
		[]byte(cfg.SessionSecret.EncryptionKey),
	)

	w.sessions = store

	stateConfig := gologin.DefaultCookieConfig
	router.Handle("/github/login", github.StateHandler(stateConfig, github.LoginHandler(oauth2Config, nil)))
	router.Handle("/github/callback", github.StateHandler(stateConfig, github.CallbackHandler(oauth2Config, w.login(), nil)))

	spa := spa.SpaHandler{
		StaticPath: "frontend/build",
		IndexPath:  "index.html",
	}
	router.PathPrefix("/").Handler(spa)

	addr := fmt.Sprintf("0.0.0.0:%s", cfg.Port)
	logger.Infof("starting web server on address %s", addr)
	server := &http.Server{
		Handler: router,
		Addr:    addr,
	}

	w.HttpServer = server
	return w, nil
}

func (s WebServer) tmpHandler(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`<!DOCTYPE html>
	<html lang="en">
	  <head>
		<title>React App</title>
	  </head>
	  <body>
		<div id="root">
		  <h1>Welcome, to use this site please login</h1>
		</div>
		<button onclick="window.location.href='/';">
		  Click Here
		</button>
	  </body>
	</html>
	`))
}

func (s WebServer) login() http.Handler {
	fn := func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		user, err := github.UserFromContext(ctx)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		session := s.sessions.New("heroku-addon")
		session.Set("user-id", *user.Email)
		if err := session.Save(w); err != nil {
			http.Error(w, "could not create session", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, req, "/", http.StatusFound)

	}
	return http.HandlerFunc(fn)
}

func (s WebServer) requireLogin(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, req *http.Request) {
		session, err := s.sessions.Get(req, "my-app")
		if err != nil {
			http.Error(w, "missing session", http.StatusUnauthorized)
			return
		}

		userID, present := session.GetOk("user-id")
		if !present {
			http.Error(w, "missing user-id", http.StatusUnauthorized)
			return
		}

		fmt.Println(userID)
		next.ServeHTTP(w, req)

	}
	return http.HandlerFunc(fn)
}

func (s WebServer) provisionHandler(w http.ResponseWriter, req *http.Request) {
	s.logger.Infof("got request for new provisioning")

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

	id, err := s.herokuClient.GetAppId(oauthResp.AccessToken)
	if err != nil {
		s.logger.Errorf("error getting app id: %s", err)
		http.Error(w, `{"error":"error provisioning","status":"failed"}`, http.StatusInternalServerError)
		return
	}

	ownerEmail, err := heroku.GetOwnerEmail(oauthResp.AccessToken, id)
	if err != nil {
		s.logger.Errorf("error getting owner email: %s", err)
		http.Error(w, `{"error":"error provisioning","status":"failed"}`, http.StatusInternalServerError)
		return
	}

	err = s.postgresClient.CreateOrUpdateAccount(s.cryptoUtil, payload.UUID, ownerEmail, oauthResp.AccessToken, oauthResp.RefreshToken)
	if err != nil {
		s.logger.Errorf("error creating account: %s", err)
		http.Error(w, `{"error":"error provisioning","status":"failed"}`, http.StatusInternalServerError)
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

func (s WebServer) deprovisionHandler(w http.ResponseWriter, req *http.Request) {
	s.logger.Infof("got request to delete addon")

	vars := gmux.Vars(req)

	s.logger.Infof("resource_uuid: %s", vars["resource_uuid"])

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
