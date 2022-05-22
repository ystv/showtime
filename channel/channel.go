package channel

import (
	"context"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/ystv/showtime/brave"
)

type (

	// MCR manages multiple channels.
	MCR struct {
		db    *sqlx.DB
		brave *brave.Braver
	}

	// Channel add redundancy to a stream.
	Channel struct {
		ID      int    `db:"channel_id" json:"channelID"`
		Title   string `db:"title" json:"title"`
		MixerID int    `db:"mixer_id"`
	}

	// NewChannel creates a new instance of a channel.
	NewChannel struct {
		Title string `db:"title" json:"title" form:"title"`
	}
)

var (
	// ErrTitleEmpty validation error when title is empty.
	ErrTitleEmpty = errors.New("title is empty")
)

// Schema represents the channel package in the database.
var Schema = `
CREATE TABLE channels (
	channel_id integer NOT NULL PRIMARY KEY AUTOINCREMENT,
	title text NOT NULL,
	mixer_id int NOT NULL
);
`

// NewMCR creates a new channel manager.
func NewMCR(db *sqlx.DB, brave *brave.Braver) *MCR {
	return &MCR{
		db:    db,
		brave: brave,
	}
}

// New creates a new channel including a mixer.
func (mcr *MCR) New(ctx context.Context, ch NewChannel) error {
	if len(ch.Title) == 0 {
		return ErrTitleEmpty
	}

	m, err := mcr.brave.NewMixer(ctx)
	if err != nil {
		return fmt.Errorf("failed to create mixer: %w", err)
	}

	_, err = mcr.brave.NewOutput(ctx, m)
	if err != nil {
		return fmt.Errorf("failed to create output: %w", err)
	}

	_, err = mcr.db.ExecContext(ctx, `
		INSERT INTO channels (title, mixer_id)
		VALUES ($1, $2);`, ch.Title, m.ID)
	if err != nil {
		return fmt.Errorf("failed to insert channel: %w", err)
	}

	return nil
}

// List retrieves a list of all channels.
func (mcr *MCR) List(ctx context.Context) ([]Channel, error) {
	ch := []Channel{}
	err := mcr.db.SelectContext(ctx, &ch, `
		SELECT channel_id, title, mixer_id
		FROM channels;
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to get list of channels: %w", err)
	}
	return ch, nil
}
