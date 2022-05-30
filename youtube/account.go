package youtube

import (
	"context"
	"fmt"

	"google.golang.org/api/youtube/v3"
)

type (
	// Account is a YouTube account that is integrated.
	Account struct {
		ID      int `db:"account_id"`
		TokenID int `db:"token_id"`
	}
)

// NewAccount adds a reference to a YouTube account that enabled integration.
func (y *YouTube) NewAccount(ctx context.Context, tokenID int) error {
	httpClient, err := y.auth.GetHTTPClient(ctx, tokenID)
	ytClient, err := youtube.New(httpClient)
	if err != nil {
		return fmt.Errorf("failed to create youtube service: %w", err)
	}

	accountID := 0
	err = y.db.GetContext(ctx, &accountID, `
		INSERT INTO youtube_accounts(token_id)
		VALUES ($1)
		RETURNING account_id;
	`, tokenID)
	if err != nil {
		return fmt.Errorf("failed to add account to store: %w", err)
	}

	y.youtubers[accountID] = newYouTuber(accountID, y.db, ytClient)
	return nil
}

// DeleteAccount removes a youtube account from ShowTime management.
func (y *YouTube) DeleteAccount(ctx context.Context, accountID int) error {
	_, err := y.db.ExecContext(ctx, `
		DELETE FROM youtube_accounts
		WHERE account_id = $1;
	`, accountID)
	if err != nil {
		return fmt.Errorf("failed to delete account from store: %w", err)
	}

	delete(y.youtubers, accountID)

	return nil
}

func (y *YouTube) listAccounts(ctx context.Context) ([]Account, error) {
	accounts := []Account{}
	err := y.db.SelectContext(ctx, &accounts, `
		SELECT account_id, token_id
		FROM youtube_accounts;`)
	if err != nil {
		return nil, fmt.Errorf("failed to list accounts in store: %w", err)
	}
	return accounts, nil
}
