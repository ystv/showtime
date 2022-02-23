package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

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

func (a *Auther) GetClient(ctx context.Context, tok *oauth2.Token) *http.Client {
	return a.config.Client(ctx, tok)
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

func (a *Auther) StoreToken(userID string, tok *oauth2.Token) error {
	f, err := os.OpenFile(userID, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("unable to open oauth token cache file: %w", err)
	}
	defer f.Close()
	err = json.NewEncoder(f).Encode(tok)
	if err != nil {
		return fmt.Errorf("failed to encode oauth token to json: %w", err)
	}
	return nil
}

func (a *Auther) GetToken(userID string) (*oauth2.Token, error) {
	f, err := os.Open(userID)
	if err != nil {
		return nil, err
	}
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	defer f.Close()
	return tok, err
}
