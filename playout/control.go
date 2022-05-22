package playout

import (
	"context"
	"fmt"
)

// End stops a playout and triggers a stop on all linked services.
func (p *Playouter) End(ctx context.Context, playoutID int) error {
	po, err := p.Get(ctx, playoutID)
	if err != nil {
		return fmt.Errorf("failed to get playout: %w", err)
	}

	if po.WebsiteLinkID != "" {

	}

	if po.YouTubeLinkID != "" {
		err = p.yt.EndBroadcast(ctx, po.YouTubeLinkID)
		if err != nil {
			return fmt.Errorf("youtube failed to end broadcast: %w", err)
		}
	}

	return nil
}
