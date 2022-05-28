package mcr

import (
	"context"
	"fmt"
)

type (
	// Playout are the individual media streams that make up a channel.
	Playout struct {
		ID           int    `db:"playout_id" json:"playoutID"`
		ChannelID    int    `db:"channel_id" json:"channelID"`
		BraveInputID int    `db:"brave_input_id" json:"braveInputID"`
		SrcType      string `db:"source_type" json:"srcType"`
		SrcURI       string `db:"source_uri" json:"srcURI"`
		Title        string `db:"title" json:"title"`
		Visibility   string `db:"visibility" json:"visibility"`
		Status       string `db:"status" json:"status"`
	}
	// NewPlayout creates a new playout on a given channel.
	NewPlayout struct {
		ChannelID  string `json:"channelID" form:"channelID"`
		SrcURI     string `json:"srcURI" form:"srcURI"`
		Title      string `json:"title" form:"title"`
		Visibility string `json:"visibility" form:"visibility"`
	}
)

// StartPlayout triggers a playout to be played on a channel.
func (mcr *MCR) StartPlayout(ctx context.Context, po Playout) error {
	ch, err := mcr.GetChannel(ctx, po.ChannelID)
	if err != nil {
		return fmt.Errorf("failed to get channel: %w", err)
	}
	err = mcr.brave.CutMixerToInput(ctx, ch.MixerID, po.BraveInputID)
	if err != nil {
		return fmt.Errorf("failed to cut mixer \"%d\" to input \"%d\": %w", ch.MixerID, po.BraveInputID, err)
	}
	return nil
}

// EndPlayout triggers a playout to stopped being played on a channel.
func (mcr *MCR) EndPlayout(ctx context.Context, po Playout) error {
	err := mcr.brave.DeleteInput(ctx, po.BraveInputID)
	if err != nil {
		return fmt.Errorf("failed to delete input: %w", err)
	}
	return nil
}

// PlayPlayoutSource triggers a playout source to be played.
//
// This allows a stream to be loaded into memory and make channel's
// switch playout's without dead-air.
func (mcr *MCR) PlayPlayoutSource(ctx context.Context, po Playout) error {
	err := mcr.brave.PlayInput(ctx, po.BraveInputID)
	if err != nil {
		return fmt.Errorf("failed to play input: %w", err)
	}
	return nil
}

// NewPlayout creates a new playout on a channel.
func (mcr *MCR) NewPlayout(ctx context.Context, po NewPlayout) (int, error) {
	if po.ChannelID == "" {
		return 0, ErrSrcURIEmpty
	}
	if po.SrcURI == "" {
		return 0, ErrSrcURIEmpty
	}
	if po.Title == "" {
		return 0, ErrTitleEmpty
	}
	if po.Visibility == "" {
		return 0, ErrVisibilityEmpty
	}

	input, err := mcr.brave.NewURIInput(ctx, po.SrcURI)
	if err != nil {
		return 0, fmt.Errorf("failed to create uri input: %w", err)
	}

	playoutID := 0
	err = mcr.db.GetContext(ctx, &playoutID, `
		INSERT INTO playouts (
			brave_input_id, channel_id, source_type, source_uri, title, visibility, status
		)
		VALUES ($1, $2, 'live', $3, $4, $5, 'scheduled')
		RETURNING playout_id;`, input.ID, po.ChannelID, po.SrcURI, po.Title, po.Visibility)
	if err != nil {
		return 0, fmt.Errorf("failed to insert playout: %w", err)
	}
	return playoutID, nil
}

// GetPlayout returns a playout.
func (mcr *MCR) GetPlayout(ctx context.Context, playoutID int) (Playout, error) {
	po := Playout{}
	err := mcr.db.GetContext(ctx, &po, `
		SELECT playout_id, brave_input_id, channel_id, source_uri, title 
		FROM playouts
		WHERE playout_id  = $1;`, playoutID)
	if err != nil {
		return Playout{}, fmt.Errorf("failed to get playout: %w", err)
	}
	return po, nil
}

// DeletePlayout removes a playout.
func (mcr *MCR) DeletePlayout(ctx context.Context, playoutID int) error {
	_, err := mcr.db.ExecContext(ctx, `
		DELETE FROM playouts
		WHERE playout_id = $1;`, playoutID)
	if err != nil {
		return fmt.Errorf("failed to delete playout: %w", err)
	}
	return nil
}
