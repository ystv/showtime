package youtube

import (
	"context"
	"fmt"

	"github.com/ystv/showtime/auth"
	"google.golang.org/api/youtube/v3"
)

type (
	YouTuber struct {
		yt *youtube.Service
	}
	Broadcast struct {
		ID        string
		Title     string
		StartTime string
		IsManaged bool
		StreamID  string
	}
	// Stream is effectively what you know as the stream key
	Stream struct {
		ID     string
		Title  string
		Status string
		Ingest Ingest
	}
	Ingest struct {
		Name    string
		Address string
	}
	NewStream struct {
		Title       string
		Description string
		FrameRate   string
		IngestType  string
		Resolution  string
	}
)

func New(auth *auth.Auther) (*YouTuber, error) {
	tok, err := auth.GetToken("me")
	if err != nil {
		return nil, fmt.Errorf("failed to get token: %w", err)
	}
	client := auth.GetClient(context.Background(), tok)
	service, err := youtube.New(client)
	if err != nil {
		return nil, fmt.Errorf("failed to create youtube service: %w", err)
	}

	return &YouTuber{
		yt: service,
	}, nil
}

func (y *YouTuber) GetBroadcasts(ctx context.Context) ([]Broadcast, error) {
	req := y.yt.LiveBroadcasts.List([]string{"id,snippet"})
	req.BroadcastStatus("all")
	ytBroadcasts, err := req.Do()
	if err != nil {
		return nil, fmt.Errorf("failed to list broadcasts: %w", err)
	}
	broadcasts := []Broadcast{}
	for _, ytBroadcast := range ytBroadcasts.Items {
		broadcast := Broadcast{
			ID:        ytBroadcast.Id,
			Title:     ytBroadcast.Snippet.Title,
			StartTime: ytBroadcast.Snippet.ScheduledStartTime,
		}
		broadcasts = append(broadcasts, broadcast)
	}
	return broadcasts, nil
}

func (y *YouTuber) EnableShowTimeForBroadcast(ctx context.Context, broadcastID string) error {
	stream, err := y.newStream(ctx)
	if err != nil {
		return fmt.Errorf("failed to create stream: %w", err)
	}

	err = y.switchStream(ctx, broadcastID, stream.ID)
	if err != nil {
		return fmt.Errorf("failed to switch stream: %w", err)
	}

	return nil
}

func (y *YouTuber) DisableShowTimeForBroadcast(ctx context.Context, broadcastID string) error {
	err := y.switchStream(ctx, broadcastID, "")
	if err != nil {
		return fmt.Errorf("failed to switch stream: %w", err)
	}
	return nil
}
