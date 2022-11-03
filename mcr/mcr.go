package mcr

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"github.com/jmoiron/sqlx"
	"github.com/ystv/showtime/engine"
)

type (
	// MCR manages multiple channels.
	MCR struct {
		baseServeURL  *url.URL
		outputAddress *url.URL
		db            *sqlx.DB
		eng           *engine.Enginer
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
func NewMCR(c *Config, db *sqlx.DB, eng *engine.Enginer) (*MCR, error) {
	baseServe, err := url.Parse(c.BaseServeURL)
	if err != nil {
		return nil, fmt.Errorf("invalid serve url: %w", err)
	}
	output, err := url.Parse(c.OutputAddress)
	if err != nil {
		return nil, fmt.Errorf("invalid output url: %w", err)
	}

	mcr := &MCR{
		baseServeURL:  baseServe,
		outputAddress: output,
		db:            db,
		eng:           eng,
	}, nil

	err = mcr.Start()
	if err != nil {
		return nil, fmt.Errorf("failed to start mcr: %w", err)
	}

	return mcr
}

// Start the MCR.
//
// We presume that existing channel engine allocations are already running. So
// we don't run any startup construction on them.
func (mcr *MCR) Start(ctx context.Context) error {
	engineHosts, err = newEngineHostStore(ctx) // Map of Brave HTTP clients.
	channels, err := ch.listActiveChannels(ctx)

	for _, ch := range channels {
		// Allocate engines to channels.
		err := eng.AllocateEngines(ch.ID, ch.MaxEngines)
		if err != nil {
			return fmt.Errorf("failed to allocate engines for ch %s: %w", ch.ID, err)
		}
		chEngines, err := ch.GetCurrentEngines(ctx)
		if err != nil {
			return fmt.Errorf("failed to get channel engines: %w", err)
		}

		// Allocate to channel's max.
		for i := range ch.MaxEngines - len(currentEngines) {
			for _, engineHost := range engineHosts {
				for _, chEngine := range chEngines {
					if engine.ID != chEngine.ID {
						err = ch.NewEngine(engineHost) // Will start engines
						if err != nil {
							return fmt.Errorf("failed to add engine: %w", err)
						}
					}
				}
			}
		}
	}

	return nil
}
