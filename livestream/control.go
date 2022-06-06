package livestream

import (
	"context"
	"fmt"
	"strconv"
)

// Start tiggers a start condition on all linked services.
func (ls *Livestreamer) Start(ctx context.Context, strm Livestream) error {
	links, err := ls.ListLinks(ctx, strm.ID)
	if err != nil {
		return fmt.Errorf("failed to list links: %w", err)
	}

	for _, link := range links {
		switch link.IntegrationType {
		case LinkMCR:
			playoutID, err := strconv.Atoi(link.IntegrationID)
			if err != nil {
				return fmt.Errorf("failed to parse string to int: %w", err)
			}
			po, err := ls.mcr.GetPlayout(ctx, playoutID)
			if err != nil {
				return fmt.Errorf("failed to get playout: %w", err)
			}

			err = ls.mcr.StartPlayout(ctx, po)
			if err != nil {
				return fmt.Errorf("mcr failed to start playout: %w", err)
			}

		case LinkYTExisting:
			broadcastID := link.IntegrationID
			b, err := ls.yt.GetBroadcastDetails(ctx, broadcastID)
			if err != nil {
				return fmt.Errorf("failed to get broadcast: %w", err)
			}
			yt, err := ls.yt.GetYouTuber(b.AccountID)
			if err != nil {
				return fmt.Errorf("failed to get youtuber: %w", err)
			}
			err = yt.StartBroadcast(ctx, broadcastID)
			if err != nil {
				return fmt.Errorf("youtube failed to start broadcast: %w", err)
			}
		}
	}

	err = ls.updateStatus(ctx, strm.ID, "stream-started")
	if err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}

	return nil
}

// End stops a playout and triggers a stop on all linked services.
func (ls *Livestreamer) End(ctx context.Context, strm Livestream) error {
	links, err := ls.ListLinks(ctx, strm.ID)
	if err != nil {
		return fmt.Errorf("failed to list links: %w", err)
	}

	for _, link := range links {
		switch link.IntegrationType {
		case LinkMCR:
			playoutID, err := strconv.Atoi(link.IntegrationID)
			if err != nil {
				return fmt.Errorf("failed to parse string to int: %w", err)
			}
			po, err := ls.mcr.GetPlayout(ctx, playoutID)
			if err != nil {
				return fmt.Errorf("failed to get playout: %w", err)
			}

			err = ls.mcr.EndPlayout(ctx, po)
			if err != nil {
				return fmt.Errorf("website failed to end playout: %w", err)
			}

		case LinkYTExisting:
			broadcastID := link.IntegrationID
			b, err := ls.yt.GetBroadcastDetails(ctx, broadcastID)
			if err != nil {
				return fmt.Errorf("failed to get broadcast: %w", err)
			}
			yt, err := ls.yt.GetYouTuber(b.AccountID)
			if err != nil {
				return fmt.Errorf("failed to get youtuber: %w", err)
			}
			err = yt.EndBroadcast(ctx, broadcastID)
			if err != nil {
				return fmt.Errorf("youtube failed to end broadcast: %w", err)
			}
		}
	}

	err = ls.updateStatus(ctx, strm.ID, "stream-ended")
	if err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}

	return nil
}
