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

// NewOutput creates an output of a mixer.
func (b *Braver) NewOutput(ctx context.Context, m Mixer) (Output, error) {
	data := struct {
		Type   string `json:"type"`
		Host   string `json:"host"`
		Source string `json:"source"`
	}{
		Type:   "tcp",
		Host:   "0.0.0.0",
		Source: fmt.Sprintf("mixer%d", m.ID),
	}

	body, err := json.Marshal(data)

	u := b.baseURL.ResolveReference(&url.URL{Path: "/api/outputs"})
	req, err := http.NewRequest(http.MethodPut, u.String(), bytes.NewBuffer(body))
	if err != nil {
		return Output{}, fmt.Errorf("%w: %w", ErrRequestFailed, err)
	}
	req.Header.Add("Accept", "application/json")

	res, err := b.c.Do(req)
	if err != nil {
		return Output{}, fmt.Errorf("failed to do request: %w", err)
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
			return Output{}, fmt.Errorf("failed to decode response: %w", err)
		}
		return Output{
			ID: resp.ID,
		}, nil
	case http.StatusBadRequest:
		resBytes, err := io.ReadAll(res.Body)
		if err != nil {
			return Output{}, fmt.Errorf("failed to decode response: %w", err)
		}
		return Output{}, fmt.Errorf("bad request: %s", string(resBytes))
	default:
		return Output{}, fmt.Errorf("unexpected HTTP response status code: %d", res.StatusCode)
	}
}

// ListOutputs lists all Brave outputs.
func (b *Braver) ListOutputs(ctx context.Context) ([]Output, error) {
	return nil, nil
}
