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
