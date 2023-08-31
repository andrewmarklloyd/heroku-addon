package web

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

func getGithubUserEmail(ctx context.Context, token oauth2.Token) (string, error) {
	ts := oauth2.StaticTokenSource(
		&token,
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	emails, _, err := client.Users.ListEmails(ctx, &github.ListOptions{
		PerPage: 10,
	})
	if err != nil {
		return "", fmt.Errorf("listing user emails: %w", err)
	}

	for _, userEmail := range emails {
		if userEmail != nil && *userEmail.Primary {
			return *userEmail.Email, nil
		}
	}

	return "", fmt.Errorf("no primary email found")
}

func (s WebServer) errorLogAndRedirect(w http.ResponseWriter, req *http.Request, logMessage, reason string) {
	s.logger.Errorf(logMessage)
	url := fmt.Sprintf("/login?reason=%s", url.QueryEscape(reason))
	http.Redirect(w, req, url, http.StatusFound)
}
