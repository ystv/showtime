package livestream

import (
	"context"
	"fmt"
)

// Start tiggers a start condition on all linked services.
func (ls *Livestreamer) Start(ctx context.Context, livestreamID int) error {
	strm, err := ls.Get(ctx, livestreamID)
	if err != nil {
		return fmt.Errorf("failed to get livestream: %w", err)
	}

	if strm.WebsiteLinkID != "" {
		// err = ls.mcr.StartPlayout(ctx, strm.WebsiteLinkID)
		// if err != nil {
		//	return fmt.Errorf("website failed to start playout: %w", err)
		// }
	}

	if strm.YouTubeLinkID != "" {
		err = ls.yt.StartBroadcast(ctx, strm.YouTubeLinkID)
		if err != nil {
			return fmt.Errorf("youtube failed to start broadcast: %w", err)
		}
	}

	return nil
}

// End stops a playout and triggers a stop on all linked services.
func (ls *Livestreamer) End(ctx context.Context, livestreamID int) error {
	strm, err := ls.Get(ctx, livestreamID)
	if err != nil {
		return fmt.Errorf("failed to get playout: %w", err)
	}

	if strm.WebsiteLinkID != "" {
		// err = ls.mcr.StartPlayout(ctx, strm.WebsiteLinkID)
		// if err != nil {
		//	return fmt.Errorf("website failed to start playout: %w", err)
	}

	if strm.YouTubeLinkID != "" {
		err = ls.yt.EndBroadcast(ctx, strm.YouTubeLinkID)
		if err != nil {
			return fmt.Errorf("youtube failed to end broadcast: %w", err)
		}
	}

	return nil
}
