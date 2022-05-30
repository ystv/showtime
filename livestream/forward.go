package livestream

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/ystv/showtime/ffmpeg"
)

// Forward a livestream to it's linked platforms.
func (ls *Livestreamer) Forward(ctx context.Context, strm ConsumeLivestream) error {
	if strm.MCRLinkID != "" {
		playoutID, err := strconv.Atoi(strm.MCRLinkID)
		if err != nil {
			return fmt.Errorf("failed to parse string to int: %w", err)
		}
		po, err := ls.mcr.GetPlayout(ctx, playoutID)
		if err != nil {
			return fmt.Errorf("failed to get playout: %w", err)
		}

		go func() {
			time.Sleep(1 * time.Second)
			err = ls.mcr.PlayPlayoutSource(ctx, po)
			if err != nil {
				log.Printf("failed to start mcr playout source: %w", err)
			}
		}()
	}

	if strm.YouTubeLinkID != "" {
		details, err := ls.yt.GetBroadcastDetails(ctx, strm.YouTubeLinkID)
		if err != nil {
			return fmt.Errorf("failed to get broadcast details: %w", err)
		}

		srcURL := ls.ingestAddress + "/" + strm.StreamKey
		dstURL := details.IngestAddress + "/" + details.IngestKey

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
