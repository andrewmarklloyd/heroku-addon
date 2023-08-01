package web

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/andrewmarklloyd/heroku-addon/internal/pkg/account"
	"github.com/google/uuid"
)

func (s WebServer) getUser(w http.ResponseWriter, req *http.Request) {
	userID, email, provenance, err := s.getUserInfo(req)
	if err != nil {
		s.logger.Errorf("getting user info: %s", err)
		http.Error(w, "could not get user", http.StatusBadRequest)
		return
	}

	fmt.Fprintf(w, `{"provenance":"%s","email":"%s","userID":"%s"}`, provenance, email, userID)
}

func (s WebServer) getInstances(w http.ResponseWriter, req *http.Request) {

	userID, _, _, err := s.getUserInfo(req)
	if err != nil {
		s.logger.Errorf("getting user info: %s", err)
		http.Error(w, "could not get user", http.StatusBadRequest)
		return
	}

	instances, err := s.postgresClient.GetInstances(userID)
	if err != nil {
		s.logger.Errorf("getting instances from postgres: %s", err)
		http.Error(w, "could not instances", http.StatusInternalServerError)
		return
	}

	iJson, err := json.Marshal(instances)
	if err != nil {
		s.logger.Errorf("marshalling instances to json: %s", err)
		http.Error(w, "could not instances", http.StatusInternalServerError)
		return
	}

	fmt.Fprint(w, string(iJson))
}

func (s WebServer) newInstance(w http.ResponseWriter, req *http.Request) {
	userID, _, provenance, err := s.getUserInfo(req)
	if err != nil {
		s.logger.Errorf("getting user info: %s", err)
		http.Error(w, "could not get user", http.StatusBadRequest)
		return
	}

	if provenance == "heroku" {
		s.logger.Errorf("heroku user cannot create instances")
		http.Error(w, `{"error":"heroku user cannot create instances"}`, http.StatusBadRequest)
		return
	}

	type instanceRequest struct {
		Name string `json:"name"`
		Plan string `json:"plan"`
	}
	var ir instanceRequest
	err = json.NewDecoder(req.Body).Decode(&ir)
	if err != nil {
		http.Error(w, `{"error":"parsing request"}`, http.StatusBadRequest)
		return
	}

	if ir.Name == "" || ir.Plan == "" {
		http.Error(w, `{"error":"name and plan are required"}`, http.StatusBadRequest)
		return
	}

	a := account.Instance{
		AccountID: userID,
		Id:        uuid.New().String(),
		Plan:      ir.Plan,
		Name:      ir.Name,
	}

	err = s.postgresClient.CreateOrUpdateInstance(a)
	if err != nil {
		s.logger.Errorf("creating instance: %s", err)
		http.Error(w, `{"error":"saving instance to database"}`, http.StatusBadRequest)
		return
	}
	fmt.Fprint(w, `{"status":"success"}`)
}

func (s WebServer) deleteInstance(w http.ResponseWriter, req *http.Request) {
	userID, _, provenance, err := s.getUserInfo(req)
	if err != nil {
		s.logger.Errorf("getting user info: %s", err)
		http.Error(w, "could not get user", http.StatusBadRequest)
		return
	}

	if provenance == "heroku" {
		s.logger.Errorf("heroku user cannot delete instances")
		http.Error(w, `{"error":"heroku user cannot delete instances"}`, http.StatusBadRequest)
		return
	}

	type instanceRequest struct {
		Id string `json:"id"`
	}
	var ir instanceRequest
	err = json.NewDecoder(req.Body).Decode(&ir)
	if err != nil {
		http.Error(w, `{"error":"parsing request"}`, http.StatusBadRequest)
		return
	}

	if ir.Id == "" {
		http.Error(w, `{"error":"id is required"}`, http.StatusBadRequest)
		return
	}

	err = s.postgresClient.DeleteInstance(userID, ir.Id)
	if err != nil {
		s.logger.Errorf("deleting instance: %s", err)
		http.Error(w, `{"error":"deleting instance"}`, http.StatusBadRequest)
		return
	}
	fmt.Fprint(w, `{"status":"success"}`)
}

func (s WebServer) getUserInfo(req *http.Request) (string, string, string, error) {
	session, err := s.sessionStore.Get(req, "heroku-addon")
	if err != nil {
		return "", "", "", fmt.Errorf("could not get session: %w", err)
	}

	userID, ok := session.GetOk("user-id")
	if !ok {
		return "", "", "", fmt.Errorf("user-id from session was not found")
	}

	email, ok := session.GetOk("user-email")
	if !ok {
		return "", "", "", fmt.Errorf("user-email from session was not found")
	}

	provenance, ok := session.GetOk("provenance")
	if !ok {
		return "", "", "", fmt.Errorf("provenance from session was not found")
	}

	return userID, email, provenance, nil
}
