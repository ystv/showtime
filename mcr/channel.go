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
		OutputURL         string `db:"url_name"`
		Width             int    `db:"res_width"`
		Height            int    `db:"res_height"`
		Title             string `db:"title"`
		MixerID           int    `db:"mixer_id"`
		ProgramInputID    int    `db:"program_input_id"`
		ContinuityInputID int    `db:"continuity_input_id"`
		ProgramOutputID   int    `db:"program_output_id"`
	}

	// NewChannel creates a new instance of a channel.
	NewChannel struct {
		Title   string `json:"title" form:"title"`
		URLName string `json:"urlName" form:"urlName"`
		Width   int
		Height  int
	}
)

var (
	// ErrURLNameEmpty when the URL name is empty.
	ErrURLNameEmpty = errors.New("url name is empty")
)

// setChannelProgram
func (mcr *MCR) setChannelProgram(ctx context.Context, channelID int, inputID int) error {
	mixerID := 0
	err := mcr.db.GetContext(ctx, &mixerID, `
		SELECT mixer_id
		FROM channels
		WHERE channel_id = $1`, channelID)
	err = mcr.brave.CutMixerToInput(ctx, mixerID, inputID)
	if err != nil {
		return fmt.Errorf("failed to cut mixer to input: %w", err)
	}
	_, err = mcr.db.ExecContext(ctx, `
		UPDATE channels
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

	o, err := mcr.brave.NewRTMPOutput(ctx, m, ch.OutputURL)
	if err != nil {
		return fmt.Errorf("failed to create output: %w", err)
	}

	_, err = mcr.db.ExecContext(ctx, `
		UPDATE channels SET
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
		UPDATE channels SET
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
func (mcr *MCR) NewChannel(ctx context.Context, ch NewChannel) (int, error) {
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
		INSERT INTO channels (
			status, title, url_name, res_width, res_height, mixer_id, program_input_id,
			continuity_input_id, program_output_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING channel_id;`, "off-air", ch.Title, ch.URLName, ch.Width, ch.Height, 0, 0, 0, 0)
	if err != nil {
		return 0, fmt.Errorf("failed to insert channel: %w", err)
	}

	return channelID, nil
}

// GetChannel returns a channel.
func (mcr *MCR) GetChannel(ctx context.Context, channelID int) (Channel, error) {
	ch := Channel{}
	err := mcr.db.GetContext(ctx, &ch, `
		SELECT channel_id, status, title, url_name, res_width, res_height, mixer_id,
					 program_input_id, continuity_input_id, program_output_id
		FROM channels
		WHERE channel_id  = $1;`, channelID)
	if err != nil {
		return Channel{}, fmt.Errorf("failed to get channel: %w", err)
	}
	ch.OutputURL = mcr.outputAddress.String() + "/ch-" + ch.OutputURL
	return ch, nil
}

// ListChannels retrieves a list of all channels.
func (mcr *MCR) ListChannels(ctx context.Context) ([]Channel, error) {
	ch := []Channel{}
	err := mcr.db.SelectContext(ctx, &ch, `
		SELECT channel_id, title, mixer_id
		FROM channels;
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to get list of channels: %w", err)
	}
	return ch, nil
}
