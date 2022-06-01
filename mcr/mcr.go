package mcr

import (
	"errors"

	"github.com/jmoiron/sqlx"
	"github.com/ystv/showtime/brave"
)

type (
	// MCR manages multiple channels.
	MCR struct {
		db    *sqlx.DB
		brave *brave.Braver
	}
)

var (
	// ErrChannelIDEmpty validation error when channel ID is empty.
	ErrChannelIDEmpty = errors.New("channel id is empty")
	// ErrSrcURIEmpty validation error when source URI is empty.
	ErrSrcURIEmpty = errors.New("source uri is empty")
	// ErrTitleEmpty validation error when title is empty.
	ErrTitleEmpty = errors.New("title is empty")
	// ErrVisibilityEmpty validation error when visibility is empty.
	ErrVisibilityEmpty = errors.New("visibility is empty")
)

// Schema represents the mcr package in the database.
var Schema = `
CREATE TABLE channels (
	channel_id integer NOT NULL PRIMARY KEY AUTOINCREMENT,
	title text NOT NULL,
	mixer_id integer NOT NULL
);

CREATE TABLE playouts (
	playout_id integer NOT NULL PRIMARY KEY AUTOINCREMENT,
	channel_id integer NOT NULL,
	brave_input_id integer NOT NULL,
	source_type text NOT NULL,
	source_uri text NOT NULL,
	status text NOT NULL,
	title text NOT NULL,
	description text NOT NULL,
	scheduled_start datetime NOT NULL,
	scheduled_end datetime NOT NULL,
	visibility text NOT NULL,
	FOREIGN KEY(channel_id) REFERENCES channels(channel_id)
);
`

// NewMCR creates a new channel manager.
func NewMCR(db *sqlx.DB, brave *brave.Braver) *MCR {
	return &MCR{
		db:    db,
		brave: brave,
	}
}
