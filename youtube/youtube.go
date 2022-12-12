package youtube

import (
	"context"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"

	"google.golang.org/api/youtube/v3"

	"github.com/ystv/showtime/auth"
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
