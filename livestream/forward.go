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

			go func() {
				time.Sleep(1 * time.Second)
				err = ls.mcr.PlayPlayoutSource(ctx, po)
				if err != nil {
					log.Printf("failed to start mcr playout source: %w", err)
				}
			}()

		case LinkYTExisting:
			details, err := ls.yt.GetBroadcastDetails(ctx, link.IntegrationID)
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

		case LinkRTMPOutput:
			customRTMPOutputID, err := strconv.Atoi(link.IntegrationID)
			if err != nil {
				return fmt.Errorf("failed to parse string to int: %w", err)
			}

			rtmpOutput, err := ls.GetRTMPOutput(ctx, customRTMPOutputID)
			if err != nil {
				return fmt.Errorf("failed to get custom rtmp output url: %w", err)
			}

			srcURL := ls.ingestAddress + "/" + strm.StreamKey

			go func() {
				time.Sleep(1 * time.Second)
				err = ffmpeg.NewForwardStream(srcURL, rtmpOutput.OutputURL)
				if err != nil {
					log.Printf("failed to forward custom rtmp stream: %+v", err)
				}
			}()
		}
	}

	return nil
}
