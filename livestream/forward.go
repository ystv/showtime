package livestream

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/ystv/showtime/ffmpeg"
)

// Forward a livestream to it's linked platforms.
func (ls *Livestreamer) Forward(ctx context.Context, strm ConsumeLivestream) error {
	if strm.MCRLinkID != "" {

	}

	if strm.YouTubeLinkID != "" {
		details, err := ls.yt.GetBroadcastDetails(ctx, strm.YouTubeLinkID)
		if err != nil {
			return fmt.Errorf("failed to get broadcast details: %w", err)
		}

		srcURL := ls.ingestAddress + "/" + strm.StreamKey
		dstURL := details.IngestAddress + "/" + details.StreamName

		go func() {
			time.Sleep(1 * time.Second)
			err = ffmpeg.NewForwardStream(srcURL, dstURL)
			if err != nil {
				log.Printf("failed to forward youtube stream: %+v", err)
			}
		}()
	}

	return nil
}
