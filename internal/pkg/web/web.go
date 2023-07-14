package web

import (
	"context"
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

type ContextKey string

const ContextProvenanceKey ContextKey = "provenance"

type WebServer struct {
	HttpServer     *http.Server
	cryptoUtil     crypto.Util
	postgresClient postgres.Client
	herokuClient   heroku.HerokuClient
	sessionStore   sessions.Store[string]
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

	// todo: make adding routes easier to see
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

	w.sessionStore = store
	router.Handle("/heroku/sso/login", http.HandlerFunc(w.herokuSSOHandler)).Methods(post)

	stateConfig := gologin.DefaultCookieConfig
	router.Handle("/github/login", github.StateHandler(stateConfig, github.LoginHandler(oauth2Config, nil)))
	router.Handle("/github/callback", github.StateHandler(stateConfig, github.CallbackHandler(oauth2Config, w.login(), nil)))
	router.Handle("/logout", w.requireLogin(w.logout()))

	router.Handle("/api/user", w.requireLogin(w.getUser())).Methods(get)

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
	ssoUser, err := s.herokuClient.ValidateSSO(req)
	if err != nil {
		s.logger.Errorf("validating heroku sso: %s", err)
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(`<!DOCTYPE html><html><h1>forbidden</h1></html>`))
		return
	}

	session := s.sessionStore.New("heroku-addon")
	session.Set("user-id", ssoUser.Email)
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
		<title>React App</title>
	  </head>
	  <body>
		<div id="root">
		  <h1>Welcome, to use this site please login</h1>
		</div>
		<button onclick="window.location.href='/github/login';">
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
			s.logger.Errorf("getting user from context: %s", err)
			http.Redirect(w, req, "/welcome", http.StatusFound)
			return
		}

		session := s.sessionStore.New("heroku-addon")
		session.Set("user-id", *user.Email)
		session.Set("provenance", "github")
		if err := session.Save(w); err != nil {
			s.logger.Errorf("saving session: %s", err)
			http.Redirect(w, req, "/welcome", http.StatusFound)
			return
		}

		http.Redirect(w, req, "/", http.StatusFound)

	}
	return http.HandlerFunc(fn)
}

func (s WebServer) logout() http.Handler {
	fn := func(w http.ResponseWriter, req *http.Request) {
		s.sessionStore.Destroy(w, "heroku-addon")
		http.Redirect(w, req, "/welcome", http.StatusFound)
	}
	return http.HandlerFunc(fn)
}

func (s WebServer) requireLogin(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		session, err := s.sessionStore.Get(r, "heroku-addon")

		if err != nil {
			// s.logger.Errorf("could not get session: %s", err)
			http.Redirect(w, r, "/welcome", http.StatusFound)
			return
		}

		provenance, ok := session.GetOk("provenance")
		if !ok {
			s.logger.Errorf("could not get provenance from context")
			http.Redirect(w, r, "/welcome", http.StatusFound)
			return
		}

		ctx := r.Context()
		req := r.WithContext(context.WithValue(ctx, ContextProvenanceKey, provenance))
		*r = *req

		_, present := session.GetOk("user-id")
		if !present {
			s.logger.Errorf("could not get user-id: %s", err)
			http.Redirect(w, req, "/welcome", http.StatusFound)
			return
		}

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
