package youtube

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"google.golang.org/api/youtube/v3"
)

type (
	YouTuber struct {
		yt *youtube.Service
	}
	Stream struct {
		Title     string
		StartTime string
	}
)

func New(client *http.Client) (*YouTuber, error) {
	service, err := youtube.New(client)
	if err != nil {
		return nil, fmt.Errorf("failed to create youtube service: %w", err)
	}

	return &YouTuber{
		yt: service,
	}, nil
}

func (y *YouTuber) GetStreams(ctx context.Context) error {
	req := y.yt.LiveBroadcasts.List([]string{"id,snippet"})
	req.BroadcastStatus("all")
	broadcasts, err := req.Do()
	if err != nil {
		return fmt.Errorf("failed to list broadcasts: %w", err)
	}
	for _, broadcast := range broadcasts.Items {
		stream := Stream{
			Title:     broadcast.Snippet.Title,
			StartTime: broadcast.Snippet.ScheduledStartTime,
		}
		log.Printf("%+v", stream)
	}
	return nil
}
