package mcr

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/fogleman/gg"
)

const lineSpacing = 30

type (
	// channelRundown is a basic summary of a channel.
	channelRundown struct {
		Title             string `db:"title"`
		Width             int    `db:"res_width"`
		Height            int    `db:"res_height"`
		MixerID           int    `db:"mixer_id"`
		ProgramInputID    int    `db:"program_input_id"`
		ContinuityInputID int    `db:"continuity_input_id"`
		Playouts          []playoutInfo
	}
	// playoutInfo is a basic summary of a playout.
	playoutInfo struct {
		Title          string    `db:"title"`
		ScheduledStart time.Time `db:"scheduled_start"`
	}
	newContinuityCardParams struct {
		X               int
		Y               int
		BackgroundPath  string
		DestinationPath string
		Title           string
		Message         string
		Playouts        []playoutInfo
	}
)

func (mcr *MCR) refreshContinuityCard(ctx context.Context, channelID int) error {
	cr, err := mcr.getChannelRundown(ctx, channelID)
	if err != nil {
		return fmt.Errorf("failed to get channel rundown: %w", err)
	}

	err = newContinuityCard(newContinuityCardParams{
		X:               cr.Width,
		Y:               cr.Height,
		BackgroundPath:  "assets/ch/0-card-bg.jpg",
		DestinationPath: fmt.Sprintf("assets/ch/%d-card-continuity.png", channelID),
		Title:           cr.Title,
		Playouts:        cr.Playouts,
	})
	if err != nil {
		return fmt.Errorf("failed to generate card: %w", err)
	}

	imgURI := mcr.baseServeURL.ResolveReference(&url.URL{Path: fmt.Sprintf("/assets/ch/%d-card-continuity.png", channelID)})
	i, err := mcr.brave.NewImageInput(ctx, imgURI.String())
	if err != nil {
		return fmt.Errorf("failed to create image input in brave: %w", err)
	}

	// If there is no currently an input or the continuity card is on, update
	// channel's program.
	if cr.ProgramInputID == 0 || cr.ProgramInputID == cr.ContinuityInputID {
		err = mcr.setChannelProgram(ctx, channelID, i.ID)
		if err != nil {
			return fmt.Errorf("failed to set channel program: %w", err)
		}
	}

	err = mcr.updateContinuityInput(ctx, channelID, i.ID)
	if err != nil {
		return fmt.Errorf("failed to update continuity input in store: %w", err)
	}

	if cr.ContinuityInputID != 0 {
		// Delete the old continuity input
		err = mcr.brave.DeleteInput(ctx, cr.ContinuityInputID)
		if err != nil {
			return fmt.Errorf("failed to delete input in brave: %w", err)
		}
	}

	return nil
}

func (mcr *MCR) getChannelRundown(ctx context.Context, channelID int) (channelRundown, error) {
	cr := channelRundown{}
	err := mcr.db.GetContext(ctx, &cr, `
		SELECT title, res_width, res_height, mixer_id, program_input_id,
		continuity_input_id
		FROM channels
		WHERE channel_id = $1;
	`, channelID)
	if err != nil {
		return channelRundown{}, fmt.Errorf("failed to get channel info: %w", err)
	}

	err = mcr.db.SelectContext(ctx, &cr.Playouts, `
		SELECT title, scheduled_start
		FROM playouts
		WHERE channel_id = $1
		AND visibility = 'public'
		AND status = 'scheduled';
	`, channelID)
	if err != nil {
		return channelRundown{}, fmt.Errorf("failed to get channel playouts: %w", err)
	}
	return cr, nil
}

func (mcr *MCR) updateContinuityInput(ctx context.Context, channelID int, inputID int) error {
	_, err := mcr.db.ExecContext(ctx, `
		UPDATE channels
			SET continuity_input_id = $1
		WHERE channel_id = $2;
	`, inputID, channelID)
	return err
}

func newContinuityCard(card newContinuityCardParams) error {
	dc := gg.NewContext(card.X, card.Y)

	if err := dc.LoadFontFace("/usr/share/fonts/truetype/ubuntu/Ubuntu-R.ttf", 96); err != nil {
		return fmt.Errorf("failed to load font face: %w", err)
	}

	im, err := gg.LoadImage(card.BackgroundPath)
	if err != nil {
		return fmt.Errorf("failed to load background image: %w", err)
	}

	dc.SetRGB(1, 1, 1)
	dc.DrawImage(im, 0, 0)
	dc.DrawStringAnchored(card.Title+" - We're not on-air right now", float64(card.X)/2, float64(card.Y)/4, 0.5, 0.5)

	if err := dc.LoadFontFace("/usr/share/fonts/truetype/ubuntu/Ubuntu-R.ttf", 50); err != nil {
		return fmt.Errorf("failed to load font face: %w", err)
	}

	if len(card.Playouts) != 0 {
		dc.DrawStringAnchored("Upcoming content", float64(card.X)/2, float64(card.Y)/2-float64(card.Y)/8, 0.5, 0.5)
	} else {
		dc.DrawStringAnchored("No content scheduled right now, check back soon!", float64(card.X)/2, float64(card.Y)/2+float64(card.Y)/8, 0.5, 0.5)
	}

	yPos := float64(card.Y) / 2

	for _, po := range card.Playouts {
		playout := po.Title
		dc.DrawStringAnchored(playout, float64(card.X)/4, yPos, 0.5, 0.5)
		yPos += lineSpacing + dc.FontHeight()
	}

	yPos = float64(card.Y) / 2

	for _, po := range card.Playouts {
		playout := po.ScheduledStart.Format("2 January - 3:04PM")
		dc.DrawStringAnchored(playout, float64(card.X)/2+float64(card.X)/4, yPos, 0.5, 0.5)
		yPos += lineSpacing + dc.FontHeight()
	}

	dc.DrawStringAnchored(card.Message, float64(card.X)/2, float64(card.Y)/2+float64(card.Y)/3, 0.5, 0.5)

	err = dc.SavePNG(card.DestinationPath)
	if err != nil {
		return fmt.Errorf("failed to save png: %w", err)
	}

	return nil
}
