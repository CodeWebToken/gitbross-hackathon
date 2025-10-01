// Copyright 2024 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package auth0

import (
	"context"
	"fmt"
	"strings"

	"golang.org/x/oauth2"

	"gogs.io/gogs/internal/auth"
)

// Provider contains configuration of an Auth0 authentication provider.
type Provider struct {
	config *Config
}

// NewProvider creates a new Auth0 authentication provider.
func NewProvider(cfg *Config) auth.Provider {
	return &Provider{
		config: cfg,
	}
}

// Authenticate performs OAuth2 authentication against Auth0.
// For Auth0 OAuth, the login parameter should be the authorization code
// and password should be the state parameter for validation.
func (p *Provider) Authenticate(code, state string) (*auth.ExternalAccount, error) {
	ctx := context.Background()
	
	// Exchange authorization code for access token
	token, err := p.config.ExchangeCode(ctx, code)
	if err != nil {
		if strings.Contains(err.Error(), "invalid_grant") {
			return nil, auth.ErrBadCredentials{Args: map[string]any{"code": code}}
		}
		return nil, fmt.Errorf("failed to exchange code for token: %w", err)
	}

	// Get user information from Auth0
	userInfo, err := p.config.GetUserInfo(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	// Extract username from email or sub
	username := userInfo.Email
	if username == "" {
		username = userInfo.Sub
	}
	
	// Remove domain from email to create username
	if strings.Contains(username, "@") {
		username = strings.Split(username, "@")[0]
	}

	return &auth.ExternalAccount{
		Login:    username,
		Name:     username,
		FullName: userInfo.Name,
		Email:    userInfo.Email,
		Location: userInfo.Locale,
		Website:  "",
		Admin:    false, // Auth0 users are not admins by default
	}, nil
}

// Config returns the underlying configuration of the Auth0 provider.
func (p *Provider) Config() any {
	return p.config
}

// HasTLS returns true since Auth0 always uses TLS.
func (*Provider) HasTLS() bool {
	return true
}

// UseTLS returns true since Auth0 always uses TLS.
func (*Provider) UseTLS() bool {
	return true
}

// SkipTLSVerify returns false since Auth0 uses valid certificates.
func (*Provider) SkipTLSVerify() bool {
	return false
}

// AuthCodeURL generates the Auth0 authorization URL for OAuth2 flow.
func (p *Provider) AuthCodeURL(state string) string {
	return p.config.AuthCodeURL(state)
}

// ExchangeCodeForToken exchanges authorization code for access token.
func (p *Provider) ExchangeCodeForToken(ctx context.Context, code string) (*oauth2.Token, error) {
	return p.config.ExchangeCode(ctx, code)
}

// GetUserInfo retrieves user information using access token.
func (p *Provider) GetUserInfo(ctx context.Context, token *oauth2.Token) (*UserInfo, error) {
	return p.config.GetUserInfo(ctx, token)
}