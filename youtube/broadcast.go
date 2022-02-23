package youtube

import (
	"context"
	"fmt"
)

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

type BroadcastDetails struct {
	BroadcastID   string `db:"broadcast_id"`
	IngestAddress string `db:"ingest_address"`
	StreamName    string `db:"stream_name"`
}

func (y *YouTuber) GetBroadcastDetails(ctx context.Context, broadcastID string) (BroadcastDetails, error) {
	details := BroadcastDetails{}
	err := y.db.GetContext(ctx, &details, `
		SELECT broadcast_id, ingest_address, stream_name
		FROM youtube_broadcasts WHERE broadcast_id = $1;
	`, broadcastID)
	if err != nil {
		return BroadcastDetails{}, fmt.Errorf("failed to get broadcast details: %w", err)
	}
	return details, nil
}

func (y *YouTuber) EndBroadcast(ctx context.Context, broadcastID string) error {
	_, err := y.yt.LiveBroadcasts.Transition("complete", broadcastID, []string{"id"}).Do()
	if err != nil {
		return fmt.Errorf("failed to transition broadcast to complete state")
	}
	return nil
}
