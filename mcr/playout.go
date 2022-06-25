package mcr

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type (
	// Playout are the individual media streams that make up a channel.
	Playout struct {
		ID             int       `db:"playout_id" json:"playoutID"`
		ChannelID      int       `db:"channel_id" json:"channelID"`
		BraveInputID   int       `db:"brave_input_id" json:"braveInputID"`
		SrcType        string    `db:"source_type" json:"srcType"`
		SrcURI         string    `db:"source_uri" json:"srcURI"`
		Status         string    `db:"status" json:"status"`
		Title          string    `db:"title" json:"title"`
		Description    string    `db:"description" json:"description"`
		ScheduledStart time.Time `db:"scheduled_start" json:"scheduledStart"`
		ScheduledEnd   time.Time `db:"scheduled_end" json:"scheduledEnd"`
		Visibility     string    `db:"visibility" json:"visibility"`
	}
	// EditPlayout creates or updates a playout on a given channel.
	EditPlayout struct {
		ChannelID      int       `json:"channelID" form:"channelID"`
		SrcURI         string    `json:"srcURI" form:"srcURI"`
		Title          string    `json:"title" form:"title"`
		Description    string    `json:"description" form:"description"`
		ScheduledStart time.Time `json:"scheduledStart" form:"scheduledStart"`
		ScheduledEnd   time.Time `json:"scheduledEnd" form:"scheduledEnd"`
		Visibility     string    `json:"visibility" form:"visibility"`
	}
)

var (
	// ErrPlayoutNotFound when a playout cannot be found.
	ErrPlayoutNotFound = errors.New("playout not found")
	// ErrSourceOnAir when a source is currently live, it cannot be removed.
	ErrSourceOnAir = errors.New("cannot remove source that is on air")
)

// StartPlayout triggers a playout to be played on a channel.
func (mcr *MCR) StartPlayout(ctx context.Context, po Playout) error {
	err := mcr.setChannelProgram(ctx, po.ChannelID, po.BraveInputID)
	if err != nil {
		return fmt.Errorf("failed to cut ch \"%d\" to input \"%d\": %w", po.ChannelID, po.BraveInputID, err)
	}

	_, err = mcr.db.ExecContext(ctx, `
		UPDATE mcr.playouts
		SET status = 'live'
		WHERE playout_id = $1;
	`, po.ID)
	if err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}

	err = mcr.refreshContinuityCard(ctx, po.ChannelID)
	if err != nil {
		return fmt.Errorf("failed to refresh continuity card: %w", err)
	}

	return nil
}

// EndPlayout triggers a playout to stopped being played on a channel.
func (mcr *MCR) EndPlayout(ctx context.Context, po Playout) error {
	// TODO: Re-approach this.

	continuityInputID := 0
	err := mcr.db.GetContext(ctx, &continuityInputID, `
		SELECT continuity_input_id
		FROM mcr.channels
		WHERE channel_id = $1;`, po.ChannelID)
	if err != nil {
		return fmt.Errorf("failed to get continuity input id: %w", err)
	}

	err = mcr.setChannelProgram(ctx, po.ChannelID, continuityInputID)
	if err != nil {
		return fmt.Errorf("failed to set channel program to continuity: %w", err)
	}

	err = mcr.brave.DeleteInput(ctx, po.BraveInputID)
	if err != nil {
		return fmt.Errorf("failed to delete input: %w", err)
	}

	_, err = mcr.db.ExecContext(ctx, `
		UPDATE mcr.playouts
		SET
			brave_input_id = 0,
			status = 'stream-ended'
		WHERE playout_id = $1;
	`, po.ID)
	if err != nil {
		return fmt.Errorf("failed to update status: %w", err)
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
func (mcr *MCR) NewPlayout(ctx context.Context, po EditPlayout) (int, error) {
	if po.ChannelID == 0 {
		return 0, ErrChannelIDInvalid
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

	input, err := mcr.brave.NewURIInput(ctx, po.SrcURI, false)
	if err != nil {
		return 0, fmt.Errorf("failed to create uri input: %w", err)
	}

	playoutID := 0
	err = mcr.db.GetContext(ctx, &playoutID, `
		INSERT INTO mcr.playouts (
			brave_input_id, channel_id, source_type, source_uri, status, title,
			description, scheduled_start, scheduled_end, visibility
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING playout_id;`,
		input.ID, po.ChannelID, "uri", po.SrcURI, "scheduled", po.Title,
		po.Description, po.ScheduledStart, po.ScheduledEnd, po.Visibility)
	if err != nil {
		return 0, fmt.Errorf("failed to insert playout: %w", err)
	}

	err = mcr.refreshContinuityCard(ctx, po.ChannelID)
	if err != nil {
		return 0, fmt.Errorf("failed to refresh continuity card: %w", err)
	}

	return playoutID, nil
}

// GetPlayout returns a playout.
func (mcr *MCR) GetPlayout(ctx context.Context, playoutID int) (Playout, error) {
	po := Playout{}
	err := mcr.db.GetContext(ctx, &po, `
		SELECT
			playout_id, brave_input_id, channel_id, source_type, source_uri, status,
			title, description, scheduled_start, scheduled_end, visibility
		FROM mcr.playouts
		WHERE playout_id  = $1;`, playoutID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrPlayoutNotFound
		}
		return Playout{}, fmt.Errorf("failed to get playout: %w", err)
	}
	return po, nil
}

// GetPlayoutsForChannel returns a list of playouts for a channel.
func (mcr *MCR) GetPlayoutsForChannel(ctx context.Context, ch Channel) ([]Playout, error) {
	po := []Playout{}
	err := mcr.db.SelectContext(ctx, &po, `
		SELECT
			playout_id, brave_input_id, channel_id, source_type, source_uri, status,
			title, description, scheduled_start, scheduled_end, visibility
		FROM mcr.playouts
		WHERE channel_id  = $1
		ORDER BY
			scheduled_start ASC,
			scheduled_end ASC;`, ch.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to list playouts: %w", err)
	}
	return po, nil
}

// UpdatePlayout updates an existing playout.
func (mcr *MCR) UpdatePlayout(ctx context.Context, playoutID int, po EditPlayout) error {
	if po.Title == "" {
		return ErrTitleEmpty
	}
	if po.Visibility == "" {
		return ErrVisibilityEmpty
	}

	oldPo, err := mcr.GetPlayout(ctx, playoutID)
	if err != nil {
		return fmt.Errorf("failed to get existing playout: %w", err)
	}

	// Defaults if none are provided.
	if po.ChannelID == 0 {
		po.ChannelID = oldPo.ChannelID
	}

	if po.SrcURI == "" {
		po.SrcURI = oldPo.SrcURI
	}

	// Check if we need to upate the playout's input
	inputID := 0

	if po.SrcURI != oldPo.SrcURI {
		if oldPo.Status == "live" {
			return ErrSourceOnAir
		}
		err = mcr.brave.DeleteInput(ctx, oldPo.BraveInputID)
		if err != nil {
			return fmt.Errorf("failed to delete input: %w", err)
		}
		i, err := mcr.brave.NewURIInput(ctx, po.SrcURI, false)
		if err != nil {
			return fmt.Errorf("failed to create new input: %w", err)
		}
		inputID = i.ID
	} else {
		inputID = oldPo.BraveInputID
	}

	_, err = mcr.db.ExecContext(ctx, `
		UPDATE mcr.playouts SET
			brave_input_id = $1,
			channel_id = $2,
			source_type = $3,
			source_uri = $4,
			title = $5,
			description = $6,
			scheduled_start = $7,
			scheduled_end = $8,
			visibility = $9
		WHERE playout_id = $10;`,
		inputID, po.ChannelID, "uri", po.SrcURI, po.Title, po.Description,
		po.ScheduledStart, po.ScheduledEnd, po.Visibility, playoutID)
	if err != nil {
		return fmt.Errorf("failed to update playout: %w", err)
	}

	err = mcr.refreshContinuityCard(ctx, po.ChannelID)
	if err != nil {
		return fmt.Errorf("failed to refresh continuity card: %w", err)
	}

	return nil
}

// DeletePlayout removes a playout.
func (mcr *MCR) DeletePlayout(ctx context.Context, playoutID int) error {
	po, err := mcr.GetPlayout(ctx, playoutID)
	if err != nil {
		return fmt.Errorf("failed to get playout: %w", err)
	}

	_, err = mcr.db.ExecContext(ctx, `
		DELETE FROM mcr.playouts
		WHERE playout_id = $1;`, po.ID)
	if err != nil {
		return fmt.Errorf("failed to delete playout from store: %w", err)
	}

	err = mcr.brave.DeleteInput(ctx, po.BraveInputID)
	if err != nil {
		return fmt.Errorf("failed to delete input from brave: %w", err)
	}

	ch, err := mcr.GetChannel(ctx, po.ChannelID)
	if err != nil {
		return fmt.Errorf("failed to get channel: %w", err)
	}

	if ch.Status == "on-air" {
		err = mcr.refreshContinuityCard(ctx, po.ChannelID)
		if err != nil {
			return fmt.Errorf("failed to refresh continuity card: %w", err)
		}
	}
	return nil
}

// PrettyDateTime formats dates to a more readable string.
func (po *Playout) PrettyDateTime(ts time.Time) string {
	if ts.After(time.Now().Add(time.Hour * 24)) {
		return ts.Format("15:04 02/01")
	}
	return ts.Format("15:04")
}
