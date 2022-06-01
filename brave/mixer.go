package brave

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// NewMixerParams are fields configuring the mixer.
type NewMixerParams struct {
	Width  int
	Height int
}

// NewMixer creates a mixer object in brave and returns it's ID.
func (b *Braver) NewMixer(ctx context.Context, p NewMixerParams) (Mixer, error) {
	data := struct {
		Pattern string `json:"pattern"`
		Width   int    `json:"width"`
		Height  int    `json:"height"`
	}{
		Pattern: "0",
		Width:   p.Width,
		Height:  p.Height,
	}

	body, err := json.Marshal(data)
	if err != nil {
		return Mixer{}, fmt.Errorf("failed to marshal json: %w", err)
	}

	u := b.baseURL.ResolveReference(&url.URL{Path: "/api/mixers"})
	req, err := http.NewRequest(http.MethodPut, u.String(), bytes.NewBuffer(body))
	if err != nil {
		return Mixer{}, fmt.Errorf("%w: %w", ErrRequestFailed, err)
	}
	req.Header.Add("Accept", "application/json")

	res, err := b.c.Do(req)
	if err != nil {
		return Mixer{}, fmt.Errorf("failed to do request: %w", err)
	}
	defer res.Body.Close()

	switch res.StatusCode {
	case http.StatusOK:
		resp := &struct {
			ID  int    `json:"id"`
			UID string `json:"uid"`
		}{}
		err = json.NewDecoder(res.Body).Decode(resp)
		if err != nil {
			return Mixer{}, fmt.Errorf("failed to decode response: %w", err)
		}
		return Mixer{
			ID: resp.ID,
		}, nil
	case http.StatusBadRequest:
		resBytes, err := io.ReadAll(res.Body)
		if err != nil {
			return Mixer{}, fmt.Errorf("failed to decode response: %w", err)
		}
		return Mixer{}, fmt.Errorf("bad request: %s", string(resBytes))
	default:
		return Mixer{}, fmt.Errorf("unexpected HTTP response status code: %d", res.StatusCode)
	}
}

// CutMixerToInput sets a mixer's program output to a given input.
func (b *Braver) CutMixerToInput(ctx context.Context, mixerID int, inputID int) error {
	data := struct {
		UID string `json:"uid"`
	}{
		UID: fmt.Sprintf("input%d", inputID),
	}

	body, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal json: %w", err)
	}

	u := b.baseURL.ResolveReference(&url.URL{Path: fmt.Sprintf("/api/mixers/%d/cut_to_source", mixerID)})
	req, err := http.NewRequest(http.MethodPost, u.String(), bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("%w: %w", ErrRequestFailed, err)
	}
	req.Header.Add("Accept", "application/json")

	res, err := b.c.Do(req)
	if err != nil {
		return fmt.Errorf("failed to do request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		resBytes, err := io.ReadAll(res.Body)
		if err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
		return fmt.Errorf("bad request: %s", string(resBytes))
	}

	return nil
}
