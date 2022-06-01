package mcr

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/jmoiron/sqlx"
	"github.com/ystv/showtime/brave"
)

type (
	// MCR manages multiple channels.
	MCR struct {
		baseServeURL *url.URL
		db           *sqlx.DB
		brave        *brave.Braver
	}
	// Config to configure Brave.
	Config struct {
		BaseServeURL string
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
	res_width integer NOT NULL,
	res_height integer NOT NULL,
	mixer_id integer NOT NULL,
	program_input_id integer NOT NULL,
	continuity_input_id integer NOT NULL
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
func NewMCR(c *Config, db *sqlx.DB, brave *brave.Braver) (*MCR, error) {
	u, err := url.Parse(c.BaseServeURL)
	if err != nil {
		return nil, fmt.Errorf("invalid serve url: %w", err)
	}
	return &MCR{
		baseServeURL: u,
		db:           db,
		brave:        brave,
	}, nil
}
