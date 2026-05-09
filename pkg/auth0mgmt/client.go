package auth0mgmt

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"
)

// Client calls the Auth0 Management API using the client_credentials grant.
// The management token is cached until 60 seconds before expiry.
type Client struct {
	domain       string
	clientID     string
	clientSecret string
	audience     string
	httpClient   *http.Client

	mu         sync.Mutex
	token      string
	tokenExpAt time.Time
}

func NewClient(domain, clientID, clientSecret string) *Client {
	return &Client{
		domain:       domain,
		clientID:     clientID,
		clientSecret: clientSecret,
		audience:     "https://" + domain + "/api/v2/",
		httpClient:   &http.Client{Timeout: 10 * time.Second},
	}
}

// SetUserRole patches app_metadata.tipo for the given Auth0 user ID.
func (c *Client) SetUserRole(ctx context.Context, auth0UserID, role string) error {
	token, err := c.accessToken(ctx)
	if err != nil {
		return fmt.Errorf("auth0mgmt: get token: %w", err)
	}

	body, _ := json.Marshal(map[string]any{
		"app_metadata": map[string]string{"tipo": role},
	})

	url := fmt.Sprintf("https://%s/api/v2/users/%s", c.domain, url.PathEscape(auth0UserID))
	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("auth0mgmt: patch user: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		var e struct {
			Message string `json:"message"`
		}
		json.NewDecoder(resp.Body).Decode(&e)
		return fmt.Errorf("auth0mgmt: patch user %s: %s", resp.Status, e.Message)
	}

	return nil
}

func (c *Client) accessToken(ctx context.Context) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.token != "" && time.Now().Before(c.tokenExpAt) {
		return c.token, nil
	}

	body, _ := json.Marshal(map[string]string{
		"grant_type":    "client_credentials",
		"client_id":     c.clientID,
		"client_secret": c.clientSecret,
		"audience":      c.audience,
	})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://"+c.domain+"/oauth/token", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
		Error       string `json:"error"`
		Description string `json:"error_description"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	if result.Error != "" {
		return "", fmt.Errorf("auth0mgmt: %s: %s", result.Error, result.Description)
	}

	c.token = result.AccessToken
	// Renova 60s antes de expirar para evitar tokens expirados em voo
	c.tokenExpAt = time.Now().Add(time.Duration(result.ExpiresIn-60) * time.Second)

	return c.token, nil
}
