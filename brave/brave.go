package brave

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
)

type (
	// Braver a Brave client.
	Braver struct {
		baseURL *url.URL
		c       *http.Client
	}
	// Config to configure Brave.
	Config struct {
		Endpoint string
	}

	// Mixer provides media mixing.
	Mixer struct {
		ID int
	}
	// Output provides output from a mixer.
	Output struct {
		ID  int
		Src string
	}
)

var (
	// ErrInvalidBaseURL when an invalid base URL is given.
	ErrInvalidBaseURL = errors.New("failed to parse base URL")
	// ErrRequestFailed to make a HTTP request.
	ErrRequestFailed = errors.New("failed to make request")
)

// New creates a new Brave client.
func New(c Config) (*Braver, error) {
	u, err := url.Parse(c.Endpoint)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidBaseURL, err)
	}
	return &Braver{
		baseURL: u,
		c:       &http.Client{},
	}, nil
}
