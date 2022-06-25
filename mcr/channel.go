package mcr

import (
	"context"
	"errors"
	"fmt"

	"github.com/ystv/showtime/brave"
)

type (
	// Channel add redundancy to a stream.
	Channel struct {
		ID                int    `db:"channel_id"`
		Status            string `db:"status"`
		URLName           string `db:"url_name"`
		OutputURL         string
		Width             int    `db:"res_width"`
		Height            int    `db:"res_height"`
		Title             string `db:"title"`
		MixerID           int    `db:"mixer_id"`
		ProgramInputID    int    `db:"program_input_id"`
		ContinuityInputID int    `db:"continuity_input_id"`
		ProgramOutputID   int    `db:"program_output_id"`
	}

	// EditChannel creates or updates a channel.
	EditChannel struct {
		Title   string `json:"title" form:"title"`
		URLName string `json:"urlName" form:"urlName"`
		Width   int
		Height  int
	}
)

var (
	// ErrURLNameEmpty when the URL name is empty.
	ErrURLNameEmpty = errors.New("url name is empty")
	// ErrChannelOnAir when the channel is on air.
	ErrChannelOnAir = errors.New("channel is on-air")
	// ErrChannelNotArchived when a channel is not in the archive status.
	ErrChannelNotArchived = errors.New("channel is not archived")
)

// setChannelProgram
func (mcr *MCR) setChannelProgram(ctx context.Context, channelID int, inputID int) error {
	mixerID := 0
	err := mcr.db.GetContext(ctx, &mixerID, `
		SELECT mixer_id
		FROM mcr.channels
		WHERE channel_id = $1`, channelID)
	err = mcr.brave.CutMixerToInput(ctx, mixerID, inputID)
	if err != nil {
		return fmt.Errorf("failed to cut mixer to input: %w", err)
	}
	_, err = mcr.db.ExecContext(ctx, `
		UPDATE mcr.channels
			SET program_input_id = $1
		WHERE channel_id = $2;
	`, inputID, channelID)
	if err != nil {
		return fmt.Errorf("failed to update program input in store: %w", err)
	}
	return nil
}

// SetChannelOnAir starts the channel's broadcast.
func (mcr *MCR) SetChannelOnAir(ctx context.Context, ch Channel) error {
	p := brave.NewMixerParams{
		Width:  ch.Width,
		Height: ch.Height,
	}

	m, err := mcr.brave.NewMixer(ctx, p)
	if err != nil {
		return fmt.Errorf("failed to create mixer: %w", err)
	}

	o, err := mcr.brave.NewRTMPOutput(ctx, m, mcr.outputAddress.String()+"/"+ch.URLName)
	if err != nil {
		return fmt.Errorf("failed to create output: %w", err)
	}

	_, err = mcr.db.ExecContext(ctx, `
		UPDATE mcr.channels SET
			status = 'on-air',
			mixer_id = $1,
			program_output_id = $2
		WHERE channel_id = $3;
	`, m.ID, o.ID, ch.ID)
	if err != nil {
		return fmt.Errorf("failed to update channel in store: %w", err)
	}

	err = mcr.refreshContinuityCard(ctx, ch.ID)
	if err != nil {
		return fmt.Errorf("failed to refresh continuity card: %w", err)
	}

	return nil
}

// SetChannelOffAir ends the channel's broadcast.
func (mcr *MCR) SetChannelOffAir(ctx context.Context, ch Channel) error {
	err := mcr.brave.DeleteOutput(ctx, ch.ProgramOutputID)
	if err != nil {
		return fmt.Errorf("failed to delete program output: %w", err)
	}
	err = mcr.brave.DeleteMixer(ctx, ch.MixerID)
	if err != nil {
		return fmt.Errorf("failed to delete mixer: %w", err)
	}

	_, err = mcr.db.ExecContext(ctx, `
		UPDATE mcr.channels SET
			status = 'off-air',
			mixer_id = 0,
			program_output_id = 0
		WHERE channel_id = $1;
	`, ch.ID)
	if err != nil {
		return fmt.Errorf("failed to delete channel in store", err)
	}

	return nil
}

// NewChannel creates a new channel including a mixer.
func (mcr *MCR) NewChannel(ctx context.Context, ch EditChannel) (int, error) {
	// Validation.
	if len(ch.Title) == 0 {
		return 0, ErrTitleEmpty
	}

	if len(ch.URLName) == 0 {
		return 0, ErrURLNameEmpty
	}

	// Sensible defaults.
	if ch.Width == 0 {
		ch.Width = 1920
	}
	if ch.Height == 0 {
		ch.Height = 1080
	}

	channelID := 0
	err := mcr.db.GetContext(ctx, &channelID, `
		INSERT INTO mcr.channels (
			status, title, url_name, res_width, res_height, mixer_id, program_input_id,
			continuity_input_id, program_output_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING channel_id;`, "off-air", ch.Title, ch.URLName, ch.Width, ch.Height, 0, 0, 0, 0)
	if err != nil {
		return 0, fmt.Errorf("failed to insert channel: %w", err)
	}

	return channelID, nil
}

// UpdateChannel updates a given channel.
//
// Certain parameters can only be changed when the channel is off-air.
func (mcr *MCR) UpdateChannel(ctx context.Context, channelID int, ch EditChannel) error {
	// Validation.
	if len(ch.Title) == 0 {
		return ErrTitleEmpty
	}

	if len(ch.URLName) == 0 {
		return ErrURLNameEmpty
	}

	oldCh, err := mcr.GetChannel(ctx, channelID)
	if err != nil {
		return fmt.Errorf("failed to get channel: %w", err)
	}

	// Defaults.
	if ch.Width == 0 {
		ch.Width = oldCh.Width
	}
	if ch.Height == 0 {
		ch.Height = oldCh.Height
	}

	if oldCh.Status == "on-air" && (ch.Width != oldCh.Width || ch.Height != oldCh.Height || ch.URLName != oldCh.URLName) {
		// This requires us to restart brave with different parameters, so need to be off-air to perform.
		return ErrChannelOnAir
	}

	_, err = mcr.db.ExecContext(ctx, `
		UPDATE mcr.channels SET
			title = $1,
			url_name = $2,
			res_width = $3,
			res_height = $4
		WHERE channel_id = $5;`, ch.Title, ch.URLName, ch.Width, ch.Height, channelID)
	if err != nil {
		return fmt.Errorf("failed to update channel: %w", err)
	}

	err = mcr.refreshContinuityCard(ctx, channelID)
	if err != nil {
		return fmt.Errorf("failed to refresh continuity card: %w", err)
	}

	return nil
}

// GetChannel returns a channel.
func (mcr *MCR) GetChannel(ctx context.Context, channelID int) (Channel, error) {
	ch := Channel{}
	err := mcr.db.GetContext(ctx, &ch, `
		SELECT channel_id, status, title, url_name, res_width, res_height, mixer_id,
					 program_input_id, continuity_input_id, program_output_id
		FROM mcr.channels
		WHERE channel_id  = $1;`, channelID)
	if err != nil {
		return Channel{}, fmt.Errorf("failed to get channel: %w", err)
	}
	// TODO: Switch to url.JoinPath when Go 1.19 is released.
	ch.OutputURL = mcr.outputAddress.String() + "/" + ch.URLName

	return ch, nil
}

// ListChannels retrieves a list of all channels.
func (mcr *MCR) ListChannels(ctx context.Context) ([]Channel, error) {
	ch := []Channel{}
	err := mcr.db.SelectContext(ctx, &ch, `
		SELECT channel_id, title, mixer_id
		FROM mcr.channels;
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to get list of channels: %w", err)
	}
	return ch, nil
}

// ArchiveChannel puts a channel into a off-state. Effectively hiding the
// channel.
func (mcr *MCR) ArchiveChannel(ctx context.Context, ch Channel) error {
	if ch.Status != "off-air" {
		return ErrChannelOnAir
	}

	_, err := mcr.db.ExecContext(ctx, `
		UPDATE mcr.channels SET
			status = 'archived'
		WHERE channel_id = $1;`, ch.ID)
	return err
}

// UnarchiveChannel restores a channel back to service.
func (mcr *MCR) UnarchiveChannel(ctx context.Context, ch Channel) error {
	if ch.Status != "archived" {
		return ErrChannelNotArchived
	}

	_, err := mcr.db.ExecContext(ctx, `
		UPDATE mcr.channels SET
			status = 'off-air'
		WHERE channel_id = $1;`, ch.ID)
	return err
}

// DeleteChannel deletes a channel including its playout's.
//
// Restricted to just archived channel's to prevent accidental deletion.
func (mcr *MCR) DeleteChannel(ctx context.Context, ch Channel) error {
	if ch.Status != "archived" {
		return ErrChannelNotArchived
	}

	playouts, err := mcr.GetPlayoutsForChannel(ctx, ch)
	if err != nil {
		return fmt.Errorf("failed to get playouts: %w", err)
	}

	for _, play := range playouts {
		err = mcr.DeletePlayout(ctx, play.ID)
		if err != nil {
			return fmt.Errorf("failed to delete playout: %w", err)
		}
	}

	_, err = mcr.db.ExecContext(ctx, `
		DELETE FROM mcr.channels
		WHERE channel_id = $1;`, ch.ID)
	if err != nil {
		return fmt.Errorf("failed to delete channel from store: %w", err)
	}
	return nil
}
