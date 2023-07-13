package heroku

import (
	"bytes"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type HerokuClient struct {
	clientSecret  string
	addonUsername string
	addonPassword string
	ssoSalt       string
}

type ConfigVars struct {
	Config []Vars `json:"config"`
}

type Vars struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

func NewHerokuClient(clientSecret, addonUsername, addonPassword, ssoSalt string) HerokuClient {
	return HerokuClient{
		clientSecret:  clientSecret,
		addonUsername: addonUsername,
		addonPassword: addonPassword,
	}
}

func (c *HerokuClient) ValidateBasicAuth(req *http.Request) bool {
	username, password, ok := req.BasicAuth()
	if !ok {
		return false
	}

	if username != c.addonUsername {
		return false
	}

	if password != c.addonPassword {
		return false
	}

	return true
}

func (c *HerokuClient) ValidateSSO(req *http.Request) error {
	// token := req.FormValue("token")
	// context_app := req.FormValue("context_app")
	// app := req.FormValue("app")
	// id := req.FormValue("id")
	// email := req.FormValue("email")
	// user_id := req.FormValue("user_id")

	timestamp := req.FormValue("timestamp")
	resourceId := req.FormValue("resource_id")
	resourceToken := req.FormValue("resource_token")

	hasher := sha1.New()
	hasher.Write([]byte(fmt.Sprintf("%s:%s:%s", resourceId, c.ssoSalt, timestamp)))
	sha := hasher.Sum(nil)

	if string(sha) != resourceToken {
		return fmt.Errorf("generated resource token did not match posted resource token")
	}

	return nil
}

func (c *HerokuClient) ExchangeToken(code string) (OauthResponse, error) {
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)

	oauthResponse, err := c.tokenRequest(data)
	if err != nil {
		return OauthResponse{}, fmt.Errorf("making auth request: %w", err)
	}

	return oauthResponse, nil
}

func (c *HerokuClient) RefreshToken(refreshToken string) (OauthResponse, error) {
	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", refreshToken)

	oauthResponse, err := c.tokenRequest(data)
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

func (c *HerokuClient) GetAppId(token string) (string, error) {
	url := fmt.Sprintf("https://api.heroku.com/addons/%s", c.addonUsername)
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

func (c *HerokuClient) authRequest(url string, data url.Values) (string, error) {
	data.Set("client_secret", c.clientSecret)

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

func (c *HerokuClient) tokenRequest(data url.Values) (OauthResponse, error) {
	data.Set("client_secret", c.clientSecret)

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
