package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/jmoiron/sqlx"
	"golang.org/x/oauth2"
)

type (
	// Auther handles integrating with oauth2.
	Auther struct {
		config *oauth2.Config
		db     *sqlx.DB
	}
	token struct {
		ID    int
		Value *oauth2.Token
	}
)

// Schema represents the auth package in the database.
var Schema = `
CREATE TABLE auth_tokens (
	token_id integer NOT NULL PRIMARY KEY AUTOINCREMENT,
	value text NOT NULL
);
`

// NewAuther creates a oauth2 handler.
func NewAuther(db *sqlx.DB, config *oauth2.Config) *Auther {
	return &Auther{
		config: config,
		db:     db,
	}
}

// GetHTTPClient returns a http client privileged to access a provider's
// platform.
func (a *Auther) GetHTTPClient(ctx context.Context, tokenID int) (*http.Client, error) {
	token, err := a.getToken(ctx, tokenID)
	if err != nil {
		return nil, fmt.Errorf("failed to get token: %w", err)
	}

	client := a.config.Client(ctx, token)

	return client, nil
}

// GetAuthCodeURL returns a URL to the identity provider's consent page.
func (a *Auther) GetAuthCodeURL(state string) string {
	// We retrieve an offline code since we want to be able to refresh
	// this token whilst the user is not online.
	return a.config.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

// NewToken converts a code from the identity provider to a refresh token
// returning an ID to retrieve it.
func (a *Auther) NewToken(ctx context.Context, code string) (int, error) {
	// Convert code from identity provider to refresh token.
	tok, err := a.config.Exchange(ctx, code)
	if err != nil {
		return 0, fmt.Errorf("failed to exchange code for refresh token: %w", err)
	}
	tokenID, err := a.storeRefreshToken(ctx, tok)
	if err != nil {
		return 0, fmt.Errorf("failed to store refresh token: %w", err)
	}
	return tokenID, nil
}

// storeRefreshToken stores a oauth2 token.
func (a *Auther) storeRefreshToken(ctx context.Context, tok *oauth2.Token) (int, error) {
	b, err := json.Marshal(tok)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal json: %w", err)
	}
	tokenID := 0
	err = a.db.GetContext(ctx, &tokenID, `
		INSERT INTO auth_tokens (value)
		VALUES ($1)
		RETURNING token_id;
	`, string(b))
	if err != nil {
		return 0, fmt.Errorf("failed to insert token: %w", err)
	}
	return tokenID, nil
}

// getToken retrives a refresh token.
func (a *Auther) getToken(ctx context.Context, tokenID int) (*oauth2.Token, error) {
	tokenString := ""
	err := a.db.GetContext(ctx, &tokenString, `
		SELECT value
		FROM auth_tokens
		WHERE token_id = $1;
	`, tokenID)
	if err != nil {
		return nil, fmt.Errorf("failed to get token: %w", err)
	}
	tok := oauth2.Token{}
	err = json.Unmarshal([]byte(tokenString), &tok)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal tokens: %w", err)
	}
	return &tok, nil
}
