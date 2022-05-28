package mcr

import (
	"context"
	"fmt"
)

type (
	// Channel add redundancy to a stream.
	Channel struct {
		ID      int    `db:"channel_id" json:"channelID"`
		Title   string `db:"title" json:"title"`
		MixerID int    `db:"mixer_id" json:"mixerID"`
	}

	// NewChannel creates a new instance of a channel.
	NewChannel struct {
		Title string `json:"title" form:"title"`
	}
)

// NewChannel creates a new channel including a mixer.
func (mcr *MCR) NewChannel(ctx context.Context, ch NewChannel) (int, error) {
	if len(ch.Title) == 0 {
		return 0, ErrTitleEmpty
	}

	m, err := mcr.brave.NewMixer(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to create mixer: %w", err)
	}

	_, err = mcr.brave.NewOutput(ctx, m)
	if err != nil {
		return 0, fmt.Errorf("failed to create output: %w", err)
	}

	channelID := 0
	err = mcr.db.GetContext(ctx, &channelID, `
		INSERT INTO channels (title, mixer_id)
		VALUES ($1, $2)
		RETURNING channel_id;`, ch.Title, m.ID)
	if err != nil {
		return 0, fmt.Errorf("failed to insert channel: %w", err)
	}

	return channelID, nil
}

// GetChannel returns a channel.
func (mcr *MCR) GetChannel(ctx context.Context, channelID int) (Channel, error) {
	ch := Channel{}
	err := mcr.db.GetContext(ctx, &ch, `
		SELECT channel_id, title, channel_id, mixer_id
		FROM channels
		WHERE channel_id  = $1;`, channelID)
	if err != nil {
		return Channel{}, fmt.Errorf("failed to get channel: %w", err)
	}
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
