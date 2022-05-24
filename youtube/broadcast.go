package youtube

import (
	"context"
	"fmt"
)

// GetBroadcasts retrieves all broadcasts on the channel.
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

// BroadcastDetails provides info on ingest and meta.
type BroadcastDetails struct {
	BroadcastID   string `db:"broadcast_id"`
	IngestAddress string `db:"ingest_address"`
	StreamName    string `db:"stream_name"`
}

// GetBroadcastDetails retrives stream ingest and name details.
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

// StartBroadcast triggers the state to be updated to "live" making it
// visible to the audience.
func (y *YouTuber) StartBroadcast(ctx context.Context, broadcastID string) error {
	_, err := y.yt.LiveBroadcasts.Transition("live", broadcastID, []string{}).Do()
	if err != nil {
		return fmt.Errorf("failed to transitoin broadcast to live state: %w", err)
	}
	return nil
}

// EndBroadcast triggers the state to be updated to "complete" making it
// marked as done.
func (y *YouTuber) EndBroadcast(ctx context.Context, broadcastID string) error {
	_, err := y.yt.LiveBroadcasts.Transition("complete", broadcastID, []string{}).Do()
	if err != nil {
		return fmt.Errorf("failed to transition broadcast to complete state: %w", err)
	}
	return nil
}
