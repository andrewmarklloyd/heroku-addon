package web

import (
	"fmt"
	"net/http"
)

func (s WebServer) getUser() http.Handler {
	fn := func(w http.ResponseWriter, req *http.Request) {
		provenance := req.Context().Value(ContextProvenanceKey)
		p, ok := provenance.(string)
		if !ok || p == "" {
			s.logger.Errorf("provenance from context is not string or was empty")
			http.Error(w, "could not get user", http.StatusBadRequest)
			return
		}
		fmt.Fprintf(w, `{"provenance":"%s"}`, p)
	}
	return http.HandlerFunc(fn)
}

func (s WebServer) getInstances() http.Handler {
	fn := func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(w, `[{"id":"a061249b-8080-4e31-8507-92e1617c7f24","plan":"free","name":"pi-sensor"},{"id":"12f7dd3c-9c16-412b-bf8a-4f65b1844b16","plan":"staging","name":"heroku-db"},{"id":"eff3d0d1-2f88-4b78-9c6e-8d7df3d5aa2b","plan":"production","name":"my-controller"}]`)
	}
	return http.HandlerFunc(fn)
}
