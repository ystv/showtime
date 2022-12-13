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
					log.Printf("failed to start mcr playout source: %v", err)
					err = ls.CreateEvent(ctx, strm.ID, EventError, EventErrorPayload{
						Err:     err.Error(),
						Context: "mcr.PlayPlayoutSource",
					})
					if err != nil {
						log.Printf("failed to log error event: %v", err)
					}
				}
			}()

		case LinkYTNew:
			err = ls.ytForward(ctx, strm.StreamKey, link.IntegrationID)
			if err != nil {
				return fmt.Errorf("failed to forward to yt-new: %w", err)
			}

		case LinkYTExisting:
			err = ls.ytForward(ctx, strm.StreamKey, link.IntegrationID)
			if err != nil {
				return fmt.Errorf("failed to forward to yt-existing: %w", err)
			}

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
				err = ffmpeg.NewForwardStream(context.Background(), srcURL, rtmpOutput.OutputURL)
				if err != nil {
					log.Printf("failed to forward custom rtmp stream: %+v", err)
					err = ls.CreateEvent(ctx, strm.ID, EventError, EventErrorPayload{
						Err:     err.Error(),
						Context: "ffmpeg.NewForwardStream",
					})
					if err != nil {
						log.Printf("failed to log error event: %v", err)
					}
				}
			}()
		}
	}

	return nil
}

func (ls *Livestreamer) ytForward(ctx context.Context, streamKey string, broadcastID string) error {
	b, err := ls.yt.GetBroadcast(ctx, broadcastID)
	if err != nil {
		return fmt.Errorf("failed to get broadcast details: %w", err)
	}

	srcURL := ls.ingestAddress + "/" + streamKey
	dstURL := b.IngestAddress + "/" + b.IngestKey

	go func() {
		time.Sleep(1 * time.Second)
		err = ffmpeg.NewForwardStream(context.Background(), srcURL, dstURL)
		if err != nil {
			log.Printf("failed to forward youtube stream: %+v", err)
		}
	}()

	return nil
}
