package auth

import (
	"context"
	"fmt"

	"golang.org/x/oauth2"
)

type (
	Auther struct {
		config *oauth2.Config
		tokens map[string]*oauth2.Token
	}
)

func NewAuther(config *oauth2.Config) *Auther {
	return &Auther{
		config: config,
		tokens: make(map[string]*oauth2.Token),
	}
}

func (a *Auther) GetAuthCodeURL(state string) string {
	// We retrieve an offline code since we want to be able to refresh
	// this token whilst the user is not online
	return a.config.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

func (a *Auther) NewToken(ctx context.Context, code string) (*oauth2.Token, error) {
	tok, err := a.config.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange token: %w", err)
	}
	return tok, nil
}

func (a *Auther) StoreToken(userID string, tok *oauth2.Token) {
	a.tokens[userID] = tok
}

func (a *Auther) GetToken(userID string) (*oauth2.Token, error) {
	tok, ok := a.tokens[userID]
	if !ok {
		return nil, fmt.Errorf("no token found")
	}
	return tok, nil
}
