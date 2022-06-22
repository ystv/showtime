package livestream

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	petname "github.com/dustinkirkland/golang-petname"
)

// ErrStreamKeyNotFound when a given stream key is not found.
var ErrStreamKeyNotFound = errors.New("stream key not found")

// GetByStreamKey retrieves a livestream by it's stream key.
func (ls *Livestreamer) GetByStreamKey(ctx context.Context, streamKey string) (ConsumeLivestream, error) {
	strm := ConsumeLivestream{}
	err := ls.db.GetContext(ctx, &strm, `
		SELECT
			livestream_id,
			stream_key
		FROM
			livestreams
		WHERE
			stream_key = $1;`, streamKey)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ConsumeLivestream{}, ErrStreamKeyNotFound
		}
		return ConsumeLivestream{}, fmt.Errorf("failed to get stream key: %w", err)
	}
	return strm, nil
}

// RefreshStreamKey rotates the stream key to a new randomly generated one.
func (ls *Livestreamer) RefreshStreamKey(ctx context.Context, livestreamID string) error {
	_, err := ls.db.ExecContext(ctx, `
		UPDATE
			livestreams SET
				stream_key = $1
		WHERE
			livestream_id = $2;`, ls.generateStreamkey(), livestreamID)
	if err != nil {
		return fmt.Errorf("failed to update stream key: %w", err)
	}
	return nil
}

func (ls *Livestreamer) generateStreamkey() string {
	return petname.Generate(3, "-")
}
