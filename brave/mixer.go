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

// NewMixer creates a mixer object in brave and returns it's ID.
func (b *Braver) NewMixer(ctx context.Context) (Mixer, error) {
	data := struct {
		Pattern string `json:"pattern"`
	}{
		Pattern: "0",
	}

	body, err := json.Marshal(data)

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
