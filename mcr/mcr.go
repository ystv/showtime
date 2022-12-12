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
