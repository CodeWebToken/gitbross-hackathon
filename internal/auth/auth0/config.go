// Copyright 2024 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package auth0

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/oauth2"
)

// Config contains configuration for Auth0 authentication provider.
type Config struct {
	// Domain is the Auth0 domain (e.g., "your-domain.auth0.com")
	Domain string
	// ClientID is the Auth0 application client ID
	ClientID string
	// ClientSecret is the Auth0 application client secret
	ClientSecret string
	// CallbackURL is the callback URL for OAuth2 flow
	CallbackURL string
	// Scopes are the OAuth2 scopes to request
	Scopes []string
}

// OAuth2Config returns the OAuth2 configuration for Auth0
func (c *Config) OAuth2Config() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     c.ClientID,
		ClientSecret: c.ClientSecret,
		RedirectURL:  c.CallbackURL,
		Scopes:       c.Scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:  fmt.Sprintf("https://%s/authorize", c.Domain),
			TokenURL: fmt.Sprintf("https://%s/oauth/token", c.Domain),
		},
	}
}

// AuthCodeURL generates the Auth0 authorization URL
func (c *Config) AuthCodeURL(state string) string {
	config := c.OAuth2Config()
	return config.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

// ExchangeCode exchanges authorization code for access token
func (c *Config) ExchangeCode(ctx context.Context, code string) (*oauth2.Token, error) {
	config := c.OAuth2Config()
	return config.Exchange(ctx, code)
}

// UserInfo represents Auth0 user information
type UserInfo struct {
	Sub           string `json:"sub"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Locale        string `json:"locale"`
	UpdatedAt     string `json:"updated_at"`
}

// GetUserInfo retrieves user information from Auth0 using access token
func (c *Config) GetUserInfo(ctx context.Context, token *oauth2.Token) (*UserInfo, error) {
	client := c.OAuth2Config().Client(ctx, token)
	
	resp, err := client.Get(fmt.Sprintf("https://%s/userinfo", c.Domain))
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get user info: status %d, body: %s", resp.StatusCode, string(body))
	}

	var userInfo UserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, fmt.Errorf("failed to decode user info: %w", err)
	}

	return &userInfo, nil
}

// ValidateConfig validates the Auth0 configuration
func (c *Config) ValidateConfig() error {
	if c.Domain == "" {
		return fmt.Errorf("Auth0 domain is required")
	}
	if c.ClientID == "" {
		return fmt.Errorf("Auth0 client ID is required")
	}
	if c.ClientSecret == "" {
		return fmt.Errorf("Auth0 client secret is required")
	}
	if c.CallbackURL == "" {
		return fmt.Errorf("Auth0 callback URL is required")
	}
	
	// Validate domain format
	if !strings.HasSuffix(c.Domain, ".auth0.com") && !strings.Contains(c.Domain, ".") {
		return fmt.Errorf("invalid Auth0 domain format")
	}
	
	// Validate callback URL format
	if _, err := url.Parse(c.CallbackURL); err != nil {
		return fmt.Errorf("invalid callback URL: %w", err)
	}
	
	return nil
}

// DefaultScopes returns the default OAuth2 scopes for Auth0
func DefaultScopes() []string {
	return []string{"openid", "profile", "email"}
}