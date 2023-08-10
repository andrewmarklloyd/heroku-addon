package heroku

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
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
		ssoSalt:       ssoSalt,
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

func (c *HerokuClient) ValidateSSO(req *http.Request) (SSOUser, error) {
	req.ParseForm()

	timestamp := req.FormValue("timestamp")
	if timestamp == "" {
		return SSOUser{}, fmt.Errorf("timestamp not found in form data")
	}
	resourceId := req.FormValue("resource_id")
	if resourceId == "" {
		return SSOUser{}, fmt.Errorf("resource_id not found in form data")
	}
	resourceToken := req.FormValue("resource_token")
	if resourceToken == "" {
		return SSOUser{}, fmt.Errorf("resource_token not found in form data")
	}

	now := time.Now()
	fiveMinAgo := now.Add(-5 * time.Minute)
	i, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return SSOUser{}, fmt.Errorf("parsing int from timestamp payload: %w", err)
	}
	tm := time.Unix(i, 0)

	if tm.Before(fiveMinAgo) {
		return SSOUser{}, fmt.Errorf("timestamp is older than 5 minutes: %w", err)
	}

	hasher := sha1.New()
	hasher.Write([]byte(fmt.Sprintf("%s:%s:%s", resourceId, c.ssoSalt, timestamp)))
	sha := hex.EncodeToString(hasher.Sum(nil))

	if string(sha) != resourceToken {
		return SSOUser{}, fmt.Errorf("generated resource token %s did not match posted resource token %s", string(sha), resourceToken)
	}

	app := req.FormValue("app")
	if app == "" {
		return SSOUser{}, fmt.Errorf("app not found in form data")
	}
	email := req.FormValue("email")
	if email == "" {
		return SSOUser{}, fmt.Errorf("email not found in form data")
	}
	userId := req.FormValue("user_id")
	if userId == "" {
		return SSOUser{}, fmt.Errorf("user_id not found in form data")
	}

	return SSOUser{
		App:    app,
		Email:  email,
		UserID: userId,
	}, nil
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

func (c *HerokuClient) GetAppAddonInfo(token string) (AddonInfo, error) {
	url := fmt.Sprintf("https://api.heroku.com/addons/%s", c.addonUsername)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return AddonInfo{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/vnd.heroku+json; version=3")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return AddonInfo{}, fmt.Errorf("performing request to heroku: %s", err)
	}

	if resp.StatusCode != http.StatusOK {
		return AddonInfo{}, fmt.Errorf("response from heroku, receieved status code: %d", resp.StatusCode)
	}

	var addonInfo AddonInfo
	err = json.NewDecoder(resp.Body).Decode(&addonInfo)
	if err != nil {
		return AddonInfo{}, fmt.Errorf("decoding token response: %w", err)
	}

	return addonInfo, nil
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
