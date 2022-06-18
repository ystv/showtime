package youtube

import (
	"context"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/ystv/showtime/auth"
	"google.golang.org/api/youtube/v3"
)

type (
	// YouTube is a client to manage accounts and youtubers.
	YouTube struct {
		db        *sqlx.DB
		auth      *auth.Auther
		youtubers map[int]*YouTuber
	}
	// YouTuber is a small client to integrate a youtuber.
	YouTuber struct {
		accountID int
		db        *sqlx.DB
		yt        *youtube.Service
	}
)

// Schema represents the youtube package in the database.
var Schema = `
CREATE SCHEMA youtube;

CREATE TABLE youtube.accounts (
	account_id bigint GENERATED ALWAYS AS IDENTITY,
	token_id integer NOT NULL,
	PRIMARY KEY(account_id),
	CONSTRAINT fk_token FOREIGN KEY(token_id) REFERENCES auth.tokens(token_id)
);

CREATE TABLE youtube.broadcasts (
	broadcast_id text NOT NULL,
	account_id bigint NOT NULL,
	ingest_address text NOT NULL,
	ingest_key text NOT NULL,
	title text NOT NULL,
	description text NOT NULL,
	scheduled_start text NOT NULL,
	scheduled_end text NOT NULL,
	visibility text NOT NULL,
	PRIMARY KEY(broadcast_id),
	CONSTRAINT fk_account FOREIGN KEY(account_id) REFERENCES youtube.accounts(account_id)
);
`
var (
	// ErrNoYouTuberFound when the youtube account cannot be found.
	ErrNoYouTuberFound = errors.New("youtuber not found")
	// ErrTokenNotFound when the token cannot be found.
	ErrTokenNotFound = errors.New("token not found")
)

// New creates an instance of a youtuber client.
func New(ctx context.Context, db *sqlx.DB, auth *auth.Auther) (*YouTube, error) {
	yt := &YouTube{
		db:        db,
		auth:      auth,
		youtubers: map[int]*YouTuber{},
	}

	accounts, err := yt.listAccounts(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list accounts: %w", err)
	}

	for _, account := range accounts {
		httpClient, err := auth.GetHTTPClient(ctx, account.TokenID)
		ytClient, err := youtube.New(httpClient)
		if err != nil {
			return nil, fmt.Errorf("failed to create youtube service: %w", err)
		}
		yt.youtubers[account.ID] = newYouTuber(account.ID, db, ytClient)
	}
	return yt, nil
}

func newYouTuber(accountID int, db *sqlx.DB, yt *youtube.Service) *YouTuber {
	return &YouTuber{
		accountID: accountID,
		db:        db,
		yt:        yt,
	}
}

// GetYouTuber fetches a youtuber.
func (y *YouTube) GetYouTuber(accountID int) (*YouTuber, error) {
	yt, ok := y.youtubers[accountID]
	if !ok {
		return nil, ErrNoYouTuberFound
	}
	return yt, nil
}
