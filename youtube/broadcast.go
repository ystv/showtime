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
