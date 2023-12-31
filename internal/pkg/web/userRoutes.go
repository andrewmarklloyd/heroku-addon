package web

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/andrewmarklloyd/heroku-addon/internal/pkg/datadog"
)

func (s WebServer) getUser(w http.ResponseWriter, req *http.Request) {
	userInfo, err := s.getUserInfo(req)
	if err != nil {
		s.logger.Errorf("getting user info: %s", err)
		http.Error(w, "could not get user", http.StatusBadRequest)
		return
	}

	uJson, err := json.Marshal(userInfo)
	if err != nil {
		s.logger.Errorf("marshalling user info to json: %s", err)
		http.Error(w, "could not get user", http.StatusInternalServerError)
		return
	}

	fmt.Fprint(w, string(uJson))
}

func (s WebServer) getInstances(w http.ResponseWriter, req *http.Request) {

	userInfo, err := s.getUserInfo(req)
	if err != nil {
		s.logger.Errorf("getting user info: %s", err)
		http.Error(w, "could not get user", http.StatusBadRequest)
		return
	}

	instances, err := s.postgresClient.GetInstances(userInfo.UserID)
	if err != nil {
		s.logger.Errorf("getting instances from postgres: %s", err)
		http.Error(w, "could not get instances", http.StatusInternalServerError)
		return
	}

	iJson, err := json.Marshal(instances)
	if err != nil {
		s.logger.Errorf("marshalling instances to json: %s", err)
		http.Error(w, "could not get instances", http.StatusInternalServerError)
		return
	}

	fmt.Fprint(w, string(iJson))
}

func (s WebServer) deleteInstance(w http.ResponseWriter, req *http.Request) {
	s.ddClient.Publish(req.Context(), datadog.CustomMetric{
		MetricName:  datadog.MetricNameDeleteInstance,
		MetricValue: 1,
		Tags: map[string]string{
			"type": "github",
		},
	})

	userInfo, err := s.getUserInfo(req)
	if err != nil {
		s.logger.Errorf("getting user info: %s", err)
		http.Error(w, "could not get user", http.StatusBadRequest)
		return
	}

	if userInfo.Provenance == "heroku" {
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

	err = s.postgresClient.DeleteInstance(userInfo.UserID, ir.Id)
	if err != nil {
		s.logger.Errorf("deleting instance: %s", err)
		http.Error(w, `{"error":"deleting instance"}`, http.StatusBadRequest)
		return
	}
	fmt.Fprint(w, `{"status":"success"}`)
}

func (s WebServer) getUserInfo(req *http.Request) (UserInfo, error) {
	session, err := s.sessionStore.Get(req, "heroku-addon")
	if err != nil {
		return UserInfo{}, fmt.Errorf("could not get session: %w", err)
	}

	userID, ok := session.GetOk("user-id")
	if !ok {
		return UserInfo{}, fmt.Errorf("user-id from session was not found")
	}

	email, ok := session.GetOk("user-email")
	if !ok {
		return UserInfo{}, fmt.Errorf("user-email from session was not found")
	}

	name, ok := session.GetOk("user-name")
	if !ok {
		return UserInfo{}, fmt.Errorf("user-name from session was not found")
	}

	provenance, ok := session.GetOk("provenance")
	if !ok {
		return UserInfo{}, fmt.Errorf("provenance from session was not found")
	}

	stripeID := ""
	if provenance != "heroku" {
		stripeID, ok = session.GetOk("stripe-id")
		if !ok {
			return UserInfo{}, fmt.Errorf("stripe-id from session was not found")
		}
	}

	return UserInfo{
		UserID:     userID,
		Email:      email,
		Name:       name,
		Provenance: provenance,
		StripeID:   stripeID,
	}, nil
}
