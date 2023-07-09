package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/andrewmarklloyd/heroku-addon/internal/pkg/crypto"
	"github.com/andrewmarklloyd/heroku-addon/internal/pkg/heroku"
	"github.com/andrewmarklloyd/heroku-addon/internal/pkg/postgres"
	"github.com/andrewmarklloyd/heroku-addon/internal/pkg/provisioner"
	"github.com/andrewmarklloyd/heroku-addon/internal/pkg/spa"

	gmux "github.com/gorilla/mux"
)

const (
	post   = "post"
	get    = "get"
	delete = "delete"
)

type WebServer struct {
	httpServer     *http.Server
	cryptoUtil     crypto.Util
	postgresClient postgres.Client
}

func newWebServer() (WebServer, error) {
	cryptoUtil, err := crypto.NewUtil(os.Getenv("ENCRYPTION_KEY"))
	if err != nil {
		return WebServer{}, fmt.Errorf("error creating crypto client: %s", err)
	}

	postgresClient, err := postgres.NewPostgresClient(os.Getenv("DATABASE_URL"))
	if err != nil {
		return WebServer{}, fmt.Errorf("error creating postgres client: %s", err)
	}

	w := WebServer{
		cryptoUtil:     cryptoUtil,
		postgresClient: postgresClient,
	}

	router := gmux.NewRouter().StrictSlash(true)

	router.Handle("/health", http.HandlerFunc(healthHandler)).Methods(get)
	router.Handle("/heroku/resources", requireHerokuAuth(http.HandlerFunc(w.provisionHandler))).Methods(post)
	router.Handle("/heroku/resources/{resource_uuid}", requireHerokuAuth(http.HandlerFunc(w.deprovisionHandler))).Methods(delete)

	spa := spa.SpaHandler{
		StaticPath: "frontend/build",
		IndexPath:  "index.html",
	}
	router.PathPrefix("/").Handler(spa)

	port := os.Getenv("PORT")
	server := &http.Server{
		Handler: router,
		Addr:    "0.0.0.0:" + port,
	}

	w.httpServer = server
	return w, nil
}

func (s WebServer) provisionHandler(w http.ResponseWriter, req *http.Request) {
	logger.Infof("got request for new provisioning")

	var payload heroku.PlanProvisionPayload
	err := json.NewDecoder(req.Body).Decode(&payload)
	if err != nil {
		logger.Errorf("Error parsing payload: %s", err)
		http.Error(w, `{"error":"Error parsing request","status":"failed"}`, http.StatusBadRequest)
		return
	}

	logger.Infof("starting provision process for %s", payload.UUID)

	oauthResp, err := heroku.ExchangeToken(payload.OauthGrant.Code)
	if err != nil {
		logger.Errorf("error exchanging token: %s", err)
		emptyOauthResp := heroku.OauthResponse{}
		if oauthResp != emptyOauthResp {
			logger.Errorf("token response: %s", oauthResp)
		}
		http.Error(w, `{"error":"error exchanging token","status":"failed"}`, http.StatusBadRequest)
		return
	}

	id, err := heroku.GetAppId(oauthResp.AccessToken)
	if err != nil {
		logger.Errorf("error getting app id: %s", err)
		http.Error(w, `{"error":"error provisioning","status":"failed"}`, http.StatusInternalServerError)
		return
	}

	ownerEmail, err := heroku.GetOwnerEmail(oauthResp.AccessToken, id)
	if err != nil {
		logger.Errorf("error getting owner email: %s", err)
		http.Error(w, `{"error":"error provisioning","status":"failed"}`, http.StatusInternalServerError)
		return
	}

	err = s.postgresClient.CreateOrUpdateAccount(s.cryptoUtil, payload.UUID, ownerEmail, oauthResp.AccessToken, oauthResp.RefreshToken)
	if err != nil {
		logger.Errorf("error creating account: %s", err)
		http.Error(w, `{"error":"error provisioning","status":"failed"}`, http.StatusInternalServerError)
		return
	}

	err = provisioner.ProvisionResource()
	if err != nil {
		logger.Errorf("error provisioning resource: %s", err)
		http.Error(w, `{"error":"error provisioning","status":"failed"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	// w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte(fmt.Sprintf(`{"id":"%s","message":"Your add-on is provisioned!","config": { "TESTING": "hello" }
	}`, payload.UUID)))
}

func (s WebServer) deprovisionHandler(w http.ResponseWriter, req *http.Request) {
	logger.Infof("got request to delete addon")

	vars := gmux.Vars(req)

	logger.Infof("resource_uuid: %s", vars["resource_uuid"])

	w.WriteHeader(http.StatusNoContent)
	w.Write([]byte(`{"id":"hello","message":"Your add-on has been deleted."}`))

}

func requireHerokuAuth(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, req *http.Request) {
		user := os.Getenv("ADDON_USERNAME")
		pass := os.Getenv("ADDON_PASSWORD")

		username, password, ok := req.BasicAuth()
		if !ok {
			logger.Errorf("basic auth from request was not ok")
			return
		}

		if username != user {
			logger.Errorf("basic auth username was not correct")
			return
		}

		if password != pass {
			logger.Errorf("basic auth password was not correct")
			return
		}

		logger.Info("successfully authenticated request")
		next.ServeHTTP(w, req)
	}
	return http.HandlerFunc(fn)
}

func healthHandler(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, `ok`)
}
