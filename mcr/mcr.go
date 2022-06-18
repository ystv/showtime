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
		baseServeURL  *url.URL
		outputAddress *url.URL
		db            *sqlx.DB
		brave         *brave.Braver
	}
	// Config to configure Brave.
	Config struct {
		BaseServeURL  string
		OutputAddress string
	}
)

var (
	// ErrChannelIDInvalid validation error when channel ID is invalid i.e. 0.
	ErrChannelIDInvalid = errors.New("channel id is invalid")
	// ErrSrcURIEmpty validation error when source URI is empty.
	ErrSrcURIEmpty = errors.New("source uri is empty")
	// ErrTitleEmpty validation error when title is empty.
	ErrTitleEmpty = errors.New("title is empty")
	// ErrVisibilityEmpty validation error when visibility is empty.
	ErrVisibilityEmpty = errors.New("visibility is empty")
)

// Schema represents the mcr package in the database.
var Schema = `
CREATE SCHEMA mcr;

CREATE TABLE mcr.channels (
	channel_id bigint GENERATED ALWAYS AS IDENTITY,
	status text NOT NULL,
	title text NOT NULL,
	url_name text NOT NULL UNIQUE,
	res_width integer NOT NULL,
	res_height integer NOT NULL,
	mixer_id integer NOT NULL,
	program_input_id integer NOT NULL,
	continuity_input_id integer NOT NULL,
	program_output_id integer NOT NULL,
	PRIMARY KEY (channel_id)
);

CREATE TABLE mcr.playouts (
	playout_id bigint GENERATED ALWAYS AS IDENTITY,
	channel_id bigint NOT NULL,
	brave_input_id integer NOT NULL,
	source_type text NOT NULL,
	source_uri text NOT NULL,
	status text NOT NULL,
	title text NOT NULL,
	description text NOT NULL,
	scheduled_start timestamptz NOT NULL,
	scheduled_end timestamptz NOT NULL,
	visibility text NOT NULL,
	PRIMARY KEY (playout_id),
	CONSTRAINT fk_channel FOREIGN KEY(channel_id) REFERENCES mcr.channels(channel_id)
);
`

// NewMCR creates a new channel manager.
func NewMCR(c *Config, db *sqlx.DB, brave *brave.Braver) (*MCR, error) {
	baseServe, err := url.Parse(c.BaseServeURL)
	if err != nil {
		return nil, fmt.Errorf("invalid serve url: %w", err)
	}
	output, err := url.Parse(c.OutputAddress)
	if err != nil {
		return nil, fmt.Errorf("invalid output url: %w", err)
	}
	return &MCR{
		baseServeURL:  baseServe,
		outputAddress: output,
		db:            db,
		brave:         brave,
	}, nil
}
