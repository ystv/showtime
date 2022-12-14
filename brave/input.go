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

// Input are sources for mixers.
type Input struct {
	ID int
}

// PlayInput triggers Brave to start playing the input.
func (b *Braver) PlayInput(ctx context.Context, inputID int) error {
	data := struct {
		State string `json:"state"`
	}{
		State: "PLAYING",
	}

	body, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal json: %w", err)
	}

	u := b.baseURL.ResolveReference(&url.URL{Path: fmt.Sprintf("/api/inputs/%d", inputID)})
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

	resp := &struct {
		Status string `json:"status"`
	}{}
	err = json.NewDecoder(res.Body).Decode(resp)
	if err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	return nil
}

// PauseInput triggers Brave to pause the input.
func (b *Braver) PauseInput(ctx context.Context, inputID int) error {
	data := struct {
		State string `json:"state"`
	}{
		State: "PAUSED",
	}

	body, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal json: %w", err)
	}

	u := b.baseURL.ResolveReference(&url.URL{Path: fmt.Sprintf("/api/inputs/%d", inputID)})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), bytes.NewBuffer(body))
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

	resp := &struct {
		Status string `json:"status"`
	}{}
	err = json.NewDecoder(res.Body).Decode(resp)
	if err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	return nil
}

// NewURIInput creates a new URI input in Brave.
//
// General-purpose input.
func (b *Braver) NewURIInput(ctx context.Context, uri string, loop bool) (Input, error) {
	data := struct {
		Type   string `json:"type"`
		State  string `json:"state"`
		URI    string `json:"uri"`
		Volume string `json:"volume"`
		Loop   bool   `json:"loop"`
	}{
		Type:   "uri",
		State:  "NULL",
		URI:    uri,
		Volume: "1.0",
		Loop:   loop,
	}
	body, err := json.Marshal(data)
	if err != nil {
		return Input{}, fmt.Errorf("failed to marshal json: %w", err)
	}

	u := b.baseURL.ResolveReference(&url.URL{Path: "/api/inputs"})
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, u.String(), bytes.NewBuffer(body))
	if err != nil {
		return Input{}, ErrRequestFailed
	}
	req.Header.Add("Accept", "application/json")

	res, err := b.c.Do(req)
	if err != nil {
		return Input{}, fmt.Errorf("failed to do request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		resBytes, err := io.ReadAll(res.Body)
		if err != nil {
			return Input{}, fmt.Errorf("failed to decode response: %w", err)
		}
		return Input{}, fmt.Errorf("bad request: %s", string(resBytes))
	}

	resp := &struct {
		ID  int    `json:"id"`
		UID string `json:"uid"`
	}{}
	err = json.NewDecoder(res.Body).Decode(resp)
	if err != nil {
		return Input{}, fmt.Errorf("failed to decode response: %w", err)
	}

	return Input{
		ID: resp.ID,
	}, nil
}

// NewImageInput creates a new image input in Brave.
func (b *Braver) NewImageInput(ctx context.Context, uri string) (Input, error) {
	data := struct {
		Type  string `json:"type"`
		State string `json:"state"`
		URI   string `json:"uri"`
	}{
		Type:  "image",
		State: "PLAYING",
		URI:   uri,
	}
	body, err := json.Marshal(data)
	if err != nil {
		return Input{}, fmt.Errorf("failed to marshal json: %w", err)
	}

	u := b.baseURL.ResolveReference(&url.URL{Path: "/api/inputs"})
	req, err := http.NewRequest(http.MethodPut, u.String(), bytes.NewBuffer(body))
	if err != nil {
		return Input{}, ErrRequestFailed
	}
	req.Header.Add("Accept", "application/json")

	res, err := b.c.Do(req)
	if err != nil {
		return Input{}, fmt.Errorf("failed to do request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		resBytes, err := io.ReadAll(res.Body)
		if err != nil {
			return Input{}, fmt.Errorf("failed to decode response: %w", err)
		}
		return Input{}, fmt.Errorf("bad request: %s", string(resBytes))
	}

	resp := &struct {
		ID  int    `json:"id"`
		UID string `json:"uid"`
	}{}
	err = json.NewDecoder(res.Body).Decode(resp)
	if err != nil {
		return Input{}, fmt.Errorf("failed to decode response: %w", err)
	}

	return Input{
		ID: resp.ID,
	}, nil
}

// DeleteInput delete an input in Brave.
func (b *Braver) DeleteInput(ctx context.Context, inputID int) error {
	u := b.baseURL.ResolveReference(&url.URL{Path: fmt.Sprintf("/api/inputs/%d", inputID)})
	req, err := http.NewRequest(http.MethodDelete, u.String(), nil)
	if err != nil {
		return ErrRequestFailed
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
