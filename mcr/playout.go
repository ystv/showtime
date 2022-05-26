package mcr

import (
	"context"
	"fmt"
)

type (
	// Playout are the individual media streams that make up a channel.
	Playout struct {
		ID         int    `db:"playout_id" json:"playoutID"`
		SrcType    string `db:"source_type" json:"srcType"`
		SrcURI     string `db:"source_uri" json:"srcURI"`
		Title      string `db:"title" json:"title"`
		Visibility string `db:"visibility" json:"visibility"`
		Status     string `db:"status" json:"status"`
	}
	// NewPlayout creates a new playout on a given channel.
	NewPlayout struct {
		ChannelID  string `json:"channelID" form:"channelID"`
		SrcURI     string `json:"srcURI" form:"srcURI"`
		Title      string `json:"title" form:"title"`
		Visibility string `json:"visibility" form:"visibility"`
	}
)

// NewPlayout creates a new playout on a channel.
func (m *MCR) NewPlayout(ctx context.Context, po NewPlayout) (int, error) {
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
	playoutID := 0
	err := m.db.GetContext(ctx, &playoutID, `
		INSERT INTO playouts (
			channel_id, source_type, source_uri, title, visibility, status
		)
		VALUES ($1, 'live', $2, $3, $4, 'scheduled')
		RETURNING playout_id;`, po.ChannelID, po.SrcURI, po.Title, po.Visibility)
	if err != nil {
		return 0, fmt.Errorf("failed to insert playout: %w", err)
	}
	return playoutID, nil
}

// GetPlayout returns a playout.
func (m *MCR) GetPlayout(ctx context.Context, playoutID int) (Playout, error) {
	po := Playout{}
	err := m.db.GetContext(ctx, &po, `
		SELECT playout_id, title 
		FROM playouts
		WHERE playout_id  = $1;`, playoutID)
	if err != nil {
		return Playout{}, fmt.Errorf("failed to get playout: %w", err)
	}
	return po, nil
}

// DeletePlayout removes a playout.
func (m *MCR) DeletePlayout(ctx context.Context, playoutID int) error {
	_, err := m.db.ExecContext(ctx, `
		DELETE FROM playouts
		WHERE playout_id = $1;`, playoutID)
	if err != nil {
		return fmt.Errorf("failed to delete playout: %w", err)
	}
	return nil
}
