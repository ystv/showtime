package playout

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/tjarratt/babble"
)

var ErrStreamKeyNotFound = errors.New("stream key not found")

func (p *Playouter) GetByStreamKey(ctx context.Context, streamKey string) (ConsumePlayout, error) {
	po := ConsumePlayout{}
	err := p.db.GetContext(ctx, &po, `
		SELECT
			stream_key,
			website_link_id,
			youtube_link_id
		FROM
			playouts
		WHERE
			stream_key = $1;`, streamKey)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ConsumePlayout{}, ErrStreamKeyNotFound
		}
		return ConsumePlayout{}, fmt.Errorf("failed to get stream key: %w", err)
	}
	return po, nil
}

func (p *Playouter) RefreshStreamkey(ctx context.Context, playoutID string) error {
	_, err := p.db.ExecContext(ctx, `
		UPDATE
			playouts SET
				stream_key = $1
		WHERE
			playout_id = $2;`, p.generateStreamkey(), playoutID)
	if err != nil {
		return fmt.Errorf("failed to update stream key: %w", err)
	}
	return nil
}

func (p *Playouter) generateStreamkey() string {
	babbler := babble.NewBabbler()
	babbler.Separator = "-"
	babbler.Count = 3
	return strings.ToLower(babbler.Babble())
}
