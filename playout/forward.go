package playout

import (
	"context"
	"fmt"

	"github.com/ystv/showtime/ffmpeg"
)

func (p *Playouter) Forward(ctx context.Context, po ConsumePlayout) error {
	if po.WebsiteLinkID != "" {

	}

	if po.YouTubeLinkID != "" {
		details, err := p.yt.GetBroadcastDetails(ctx, po.YouTubeLinkID)
		if err != nil {
			return fmt.Errorf("failed to get broadcast details: %w", err)
		}

		srcUrl := p.ingestAddress + "/" + po.StreamKey
		dstUrl := details.IngestAddress + "/" + details.StreamName

		err = ffmpeg.NewForwardStream(srcUrl, dstUrl)
		if err != nil {
			return fmt.Errorf("failed to forward youtube stream: %w", err)
		}
	}

	return nil
}
