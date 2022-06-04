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

// Output provides output from a mixer.
type Output struct {
	ID  int
	Src string
	Dst string
}

// NewRTMPOutput creates an RTMP output of a mixer.
func (b *Braver) NewRTMPOutput(ctx context.Context, m Mixer, outURI string) (Output, error) {
	source := fmt.Sprintf("mixer%d", m.ID)
	data := struct {
		Type   string `json:"type"`
		URI    string `json:"uri"`
		Width  int    `json:"width"`
		Height int    `json:"height"`
		Source string `json:"source"`
	}{
		Type:   "rtmp",
		URI:    outURI,
		Width:  m.width,
		Height: m.height,
		Source: source,
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
			ID:  resp.ID,
			Src: source,
			Dst: outURI,
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

// NewTCPOutput creates a TCP output of a mixer.
func (b *Braver) NewTCPOutput(ctx context.Context, m Mixer, port int) (Output, error) {
	source := fmt.Sprintf("mixer%d", m.ID)
	data := struct {
		Type   string `json:"type"`
		Host   string `json:"host"`
		Port   int    `json:"port"`
		Width  int    `json:"width"`
		Height int    `json:"height"`
		Source string `json:"source"`
	}{
		Type:   "tcp",
		Host:   "0.0.0.0",
		Port:   port,
		Width:  m.width,
		Height: m.height,
		Source: source,
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
			ID:  resp.ID,
			Src: source,
			Dst: fmt.Sprintf("tcp://%s:%d", b.baseURL, port),
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
