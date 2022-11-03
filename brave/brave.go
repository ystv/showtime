package brave

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type (
	// Brave a service of brave servers.
	Brave struct {
		conf    *Config
		servers map[string]*Braver
	}
	// Braver a Brave client.
	Braver struct {
		baseURL *url.URL
		c       *http.Client
		state   *state
	}
	// Config to configure Brave.
	Config struct {
		Servers []string `json:"servers"`
	}
	state struct {
		mixers []Mixer `json:"mixers"`
	}
)

var (
	// ErrRequestFailed to make a HTTP request.
	ErrRequestFailed = errors.New("failed to make request")
)

// New creates a new Brave service client.
func New(c *Config) (*Brave, error) {
	b := &Brave{
		conf:    c,
		servers: map[string]*Braver{},
	}

	for _, server := range b.conf.Servers {
		brave, err := newBraver(server)
		if err != nil {
			return nil, fmt.Errorf("failed to create brave client: %w", err)
		}
		b.servers[server] = brave
	}

	return b, nil
}

// newBraver creates a client to a Brave server.
func newBraver(server string) (*Braver, error) {
	baseURL, err := url.Parse(server)
	if err != nil {
		return nil, fmt.Errorf("failed to parse server url: %w", err)
	}
	return &Braver{
		baseURL: baseURL,
		c:       &http.Client{},
	}, nil
}

func (b *Braver) refreshState(ctx context.Context) error {
	u := b.baseURL.ResolveReference(&url.URL{Path: "/api/all"})
	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrRequestFailed, err)
	}
	req.Header.Add("Accept", "application/json")

	res, err := b.c.Do(req)
	if err != nil {
		return fmt.Errorf("failed to do request: %w", err)
	}
	defer res.Body.Close()

	switch res.StatusCode {
	case http.StatusOK:
		state := &state{}
		err = json.NewDecoder(res.Body).Decode(state)
		if err != nil {
			return fmt.Errorf("failed to decode respone: %w", err)
		}
		b.state = state
	case http.StatusBadRequest:
		resBytes, err := io.ReadAll(res.Body)
		if err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
		return fmt.Errorf("bad request: %s", string(resBytes))
	default:
		return fmt.Errorf("unexpected HTTP response status code: %d", res.StatusCode)
	}

	return nil
}
