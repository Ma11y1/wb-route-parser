package sheets

import (
	"context"
	"encoding/json"
	"errors"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
	"net/http"
	"os"
)

type Client struct {
	httpClient *http.Client
	service    *sheets.Service
	token      *oauth2.Token
	config     *oauth2.Config
	isAuth     bool
}

func NewClient() *Client {
	return &Client{
		httpClient: nil,
		service:    nil,
		token:      nil,
		config:     nil,
		isAuth:     false,
	}
}

func (c *Client) AuthByToken(ctx context.Context, token *oauth2.Token) error {
	if token == nil {
		return errors.New("token cannot be nil\n")
	}

	c.token = token

	if c.httpClient == nil {
		c.httpClient = oauth2.NewClient(ctx, oauth2.StaticTokenSource(c.token))
	}

	var err error
	c.service, err = sheets.NewService(ctx, option.WithHTTPClient(c.httpClient))
	if err != nil {
		return err
	}

	c.isAuth = true
	return nil
}

func (c *Client) AuthByCredentials(ctx context.Context, credentials string, scope Scope, authHandle func(string) string) error {
	file, err := os.ReadFile(credentials)
	if err != nil {
		return err
	}

	c.config, err = google.ConfigFromJSON(file, string(scope))
	if err != nil {
		return err
	}

	authURL := c.config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	authCode := authHandle(authURL)
	if authCode == "" {
		return errors.New("failed to get auth code\n")
	}

	c.token, err = c.config.Exchange(ctx, authCode)
	if err != nil {
		return err
	}

	return c.AuthByToken(ctx, c.token)
}

func (c *Client) AuthByJSONTokenAutoRefresh(ctx context.Context, tokenPath string, credentialsPath string, scope Scope) error {
	tokenFile, err := os.ReadFile(tokenPath)
	if err != nil {
		return err
	}

	credentialsFile, err := os.ReadFile(credentialsPath)
	if err != nil {
		return err
	}

	c.config, err = google.ConfigFromJSON(credentialsFile, string(scope))
	if err != nil {
		return err
	}

	c.token = &oauth2.Token{}
	err = json.Unmarshal(tokenFile, &c.token)
	if err != nil {
		return err
	}

	c.httpClient = c.config.Client(ctx, c.token)

	return c.AuthByToken(ctx, c.token)
}

func (c *Client) AuthByJSONToken(ctx context.Context, tokenPath string) error {
	tokenFile, err := os.ReadFile(tokenPath)
	if err != nil {
		return err
	}

	c.token = &oauth2.Token{}
	err = json.Unmarshal(tokenFile, &c.token)
	if err != nil {
		return err
	}

	c.httpClient = oauth2.NewClient(ctx, oauth2.StaticTokenSource(c.token))

	return c.AuthByToken(ctx, c.token)
}

func (c *Client) RefreshToken(ctx context.Context) error {
	if !c.isAuth {
		return errors.New("client is not authorized\n")
	}

	if c.config == nil {
		return errors.New("invalid config client\n")
	}

	token := &oauth2.Token{
		RefreshToken: c.token.RefreshToken,
	}

	tokenSource := c.config.TokenSource(ctx, token)

	var err error
	c.token, err = tokenSource.Token()
	if err != nil {
		return errors.New("error getting new token:\n" + err.Error())
	}

	return nil
}

func (c *Client) CheckValidToken() (bool, error) {
	_, err := c.service.Spreadsheets.Get("spreadsheetId").Do()
	if err != nil {
		if apiErr, ok := err.(*googleapi.Error); ok {
			if apiErr.Code == http.StatusUnauthorized {
				return false, nil
			}
		}

		return false, errors.New("failed to check token validity:\n" + err.Error())
	}

	return true, nil
}

func (c *Client) SaveJSONToken(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	return json.NewEncoder(file).Encode(c.token)
}

func (c *Client) Internal() *sheets.Service {
	return c.service
}

func (c *Client) IsAuth() bool {
	return c.isAuth
}

func (c *Client) Close() {
	if c.httpClient != nil {
		if transport, ok := c.httpClient.Transport.(*http.Transport); ok {
			transport.CloseIdleConnections()
		}
	}

	c.httpClient = nil
	c.service = nil
	c.token = nil
	c.config = nil
	c.isAuth = false
}
