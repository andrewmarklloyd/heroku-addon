package heroku

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

type ConfigVars struct {
	Config []Vars `json:"config"`
}

type Vars struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

func ExchangeToken(code string) (OauthResponse, error) {
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)

	oauthResponse, err := tokenRequest(data)
	if err != nil {
		return OauthResponse{}, fmt.Errorf("making auth request: %w", err)
	}

	return oauthResponse, nil
}

func RefreshToken(refreshToken string) (OauthResponse, error) {
	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", refreshToken)

	oauthResponse, err := tokenRequest(data)
	if err != nil {
		return OauthResponse{}, fmt.Errorf("making auth request: %w", err)
	}

	return oauthResponse, nil
}

func GetAddonInfo(token, resourceUUID string) error {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("https://api.heroku.com/addons/%s", resourceUUID), nil)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/vnd.heroku+json; version=3")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("performing request to heroku: %s", err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response body from heroku: %s", err)
	}

	fmt.Println(string(body))

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("response from heroku, receieved status code: %d, response: %s", resp.StatusCode, string(body))
	}

	return nil
}

func GetAppId(token string) (string, error) {
	url := fmt.Sprintf("https://api.heroku.com/addons/%s", os.Getenv("ADDON_USERNAME"))
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/vnd.heroku+json; version=3")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("performing request to heroku: %s", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("response from heroku, receieved status code: %d", resp.StatusCode)
	}

	var addonInfo AddonInfo
	err = json.NewDecoder(resp.Body).Decode(&addonInfo)
	if err != nil {
		return "", fmt.Errorf("decoding token response: %w", err)
	}

	return addonInfo.App.Id, nil
}

func GetOwnerEmail(token, appId string) (string, error) {
	url := fmt.Sprintf("https://api.heroku.com/apps/%s/collaborators", appId)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/vnd.heroku+json; version=3")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("performing request to heroku: %s", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("response from heroku, receieved status code: %d", resp.StatusCode)
	}

	var appCollaborators []AppCollaborator
	err = json.NewDecoder(resp.Body).Decode(&appCollaborators)
	if err != nil {
		return "", fmt.Errorf("decoding token response: %w", err)
	}

	if len(appCollaborators) == 0 {
		return "", fmt.Errorf("did not find any collaborators")
	}

	for _, c := range appCollaborators {
		if c.Role == "owner" {
			return c.User.Email, nil
		}
	}

	return "", fmt.Errorf("did not find owner")
}

func UpdateConfigVars(token, resourceUUID string, configVars ConfigVars) error {
	j, err := json.Marshal(configVars)
	if err != nil {
		return fmt.Errorf("marshalling heartbeat: %s", err)
	}

	req, err := http.NewRequest(http.MethodPatch, fmt.Sprintf("https://api.heroku.com/addons/%s/config", resourceUUID), bytes.NewBuffer(j))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/vnd.heroku+json; version=3")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("performing request to heroku: %s", err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response body from heroku: %s", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("response from heroku, receieved status code: %d, response: %s", resp.StatusCode, string(body))
	}

	return nil
}

func authRequest(url string, data url.Values) (string, error) {
	data.Set("client_secret", os.Getenv("CLIENT_SECRET"))

	client := &http.Client{}
	r, _ := http.NewRequest(http.MethodPost, url, strings.NewReader(data.Encode()))
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(r)
	if err != nil {
		return "", fmt.Errorf("making token request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("received non 200 status code %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("reading response body: %s", err)
	}

	return string(body), nil
}

func tokenRequest(data url.Values) (OauthResponse, error) {
	data.Set("client_secret", os.Getenv("CLIENT_SECRET"))

	client := &http.Client{}
	r, _ := http.NewRequest(http.MethodPost, "https://id.heroku.com/oauth/token", strings.NewReader(data.Encode()))
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(r)
	if err != nil {
		return OauthResponse{}, fmt.Errorf("making token request: %w", err)
	}

	var oauthResponse OauthResponse
	err = json.NewDecoder(resp.Body).Decode(&oauthResponse)
	if err != nil {
		return OauthResponse{}, fmt.Errorf("decoding token response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return OauthResponse{}, fmt.Errorf("received non 200 status code %d", resp.StatusCode)
	}

	return oauthResponse, nil
}
