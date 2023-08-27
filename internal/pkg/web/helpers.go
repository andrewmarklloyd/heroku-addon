package web

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

func getGithubUserEmail(ctx context.Context, token string) (string, error) {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)
	user, _, err := client.Users.Get(ctx, "")
	if err != nil {
		return "", err
	}

	return user.GetEmail(), nil
}

func (s WebServer) errorLogAndRedirect(w http.ResponseWriter, req *http.Request, logMessage, reason string) {
	s.logger.Errorf(logMessage)
	url := fmt.Sprintf("/login?reason=%s", url.QueryEscape(reason))
	http.Redirect(w, req, url, http.StatusFound)
}
