package youtube

import (
	"context"
	"fmt"
	"time"
)

type (
	// Broadcast are the livestreams which are watchable as videos.
	Broadcast struct {
		ID        string `json:"id"`
		Title     string `json:"title"`
		StartTime string `json:"startTime"`
	}
	// BroadcastDetails provides info on ingest and meta.
	BroadcastDetails struct {
		BroadcastID   string `db:"broadcast_id"`
		AccountID     int    `db:"account_id"`
		IngestAddress string `db:"ingest_address"`
		StreamName    string `db:"stream_name"`
	}
)

// StartBroadcast triggers the state to be updated to "live" making it
// visible to the audience.
func (y *YouTuber) StartBroadcast(ctx context.Context, broadcastID string) error {
	// Check broadcast does exist.
	_, err := y.GetBroadcastDetails(ctx, broadcastID)
	if err != nil {
		return fmt.Errorf("failed to get broadcast details: %w", err)
	}
	_, err = y.yt.LiveBroadcasts.Transition("live", broadcastID, []string{}).Do()
	if err != nil {
		return fmt.Errorf("failed to transition broadcast to live state: %w", err)
	}
	return nil
}

// EndBroadcast triggers the state to be updated to "complete" making it
// marked as done.
func (y *YouTuber) EndBroadcast(ctx context.Context, broadcastID string) error {
	// Check broadcast does exist.
	_, err := y.GetBroadcastDetails(ctx, broadcastID)
	if err != nil {
		return fmt.Errorf("failed to get broadcast details: %w", err)
	}
	_, err = y.yt.LiveBroadcasts.Transition("complete", broadcastID, []string{}).Do()
	if err != nil {
		return fmt.Errorf("failed to transition broadcast to complete state: %w", err)
	}
	return nil
}

// NewExistingBroadcast enables ShowTime! integration on an existing broadcast.
//
// The incoming video stream will be forwarded to YouTube and the state can be
// controlled.
func (y *YouTuber) NewExistingBroadcast(ctx context.Context, broadcastID string) error {
	stream, err := y.newStream(ctx)
	if err != nil {
		return fmt.Errorf("failed to create stream: %w", err)
	}

	err = y.switchStream(ctx, broadcastID, stream.ID)
	if err != nil {
		return fmt.Errorf("failed to switch stream: %w", err)
	}

	_, err = y.db.ExecContext(ctx, `
		INSERT INTO youtube_broadcasts (
			broadcast_id,
			account_id,
			ingest_address,
			ingest_key
		) VALUES ($1, $2, $3, $4);`, broadcastID, y.accountID, stream.Ingest.Address, stream.Ingest.Key)
	if err != nil {
		return fmt.Errorf("failed to add broadcast to store: %w", err)
	}

	return nil
}

// DeleteExistingBroadcast disables ShowTime! integration and leaves the
// broadcast to still exist on YouTube.
func (y *YouTube) DeleteExistingBroadcast(ctx context.Context, broadcastID string) error {
	b, err := y.GetBroadcastDetails(ctx, broadcastID)
	if err != nil {
		return fmt.Errorf("failed to get broadcast details: %w", err)
	}
	yt, err := y.GetYouTuber(b.AccountID)
	if err != nil {
		return fmt.Errorf("failed to get youtuber: %w", err)
	}
	return yt.DeleteExistingBroadcast(ctx, broadcastID)
}

// DeleteExistingBroadcast disables ShowTime! integration and leaves the
// broadcast to still exist on YouTube.
//
// Deletes the generated stream key to the default and updates the database to
// de-link the broadcast.
func (y *YouTuber) DeleteExistingBroadcast(ctx context.Context, broadcastID string) error {
	// Check broadcast exists.
	_, err := y.GetBroadcastDetails(ctx, broadcastID)
	if err != nil {
		return fmt.Errorf("failed to get broadcast details: %w", err)
	}
	err = y.switchStream(ctx, broadcastID, "")
	if err != nil {
		return fmt.Errorf("failed to switch stream: %w", err)
	}

	_, err = y.db.ExecContext(ctx, `
		DELETE FROM youtube_broadcasts
		WHERE broadcast_id = $1;`, broadcastID)
	if err != nil {
		return fmt.Errorf("failed to delete broadcast: %w", err)
	}

	return nil
}

// ListBroadcasts retrieves all broadcasts on all youtubers.
//
// Retrieves directly from YouTube so slightly slow.
func (y *YouTube) ListBroadcasts(ctx context.Context) ([]Broadcast, error) {
	broadcasts := []Broadcast{}
	for _, yt := range y.youtubers {
		b, err := yt.ListBroadcasts(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get youtubers broadcasts: %w", err)
		}
		broadcasts = append(broadcasts, b...)
	}
	return broadcasts, nil
}

// ListBroadcasts retrieves all broadcasts on the channel.
//
// Retrieves directly from YouTube so slightly slow.
func (y *YouTuber) ListBroadcasts(ctx context.Context) ([]Broadcast, error) {
	broadcasts := []Broadcast{}
	req := y.yt.LiveBroadcasts.List([]string{"id,snippet"})
	req.BroadcastStatus("all")
	ytBroadcasts, err := req.Do()
	if err != nil {
		return nil, fmt.Errorf("failed to list broadcasts: %w", err)
	}
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

// GetBroadcastDetails retrives stream ingest and name details.
//
// Retrieved from store which is quick.
func (y *YouTube) GetBroadcastDetails(ctx context.Context, broadcastID string) (BroadcastDetails, error) {
	details := BroadcastDetails{}
	err := y.db.GetContext(ctx, &details, `
		SELECT broadcast_id, account_id, ingest_address, stream_name
		FROM youtube_broadcasts
		WHERE broadcast_id = $1;
	`, broadcastID)
	if err != nil {
		return BroadcastDetails{}, fmt.Errorf("failed to get broadcast details: %w", err)
	}
	return details, nil
}

// GetBroadcastDetails retrives stream ingest and name details.
//
// Retrieved from store which is quick.
func (y *YouTuber) GetBroadcastDetails(ctx context.Context, broadcastID string) (BroadcastDetails, error) {
	details := BroadcastDetails{}
	err := y.db.GetContext(ctx, &details, `
		SELECT broadcast_id, account_id, ingest_address, stream_name
		FROM youtube_broadcasts
		WHERE broadcast_id = $1;
	`, broadcastID)
	if err != nil {
		return BroadcastDetails{}, fmt.Errorf("failed to get broadcast details: %w", err)
	}
	return details, nil
}

// PrettyDateTime formats dates to a more readable string.
func (b *Broadcast) PrettyDateTime() string {
	ts, err := time.Parse(time.RFC3339, b.StartTime)
	if err != nil {
		return err.Error()
	}

	return ts.Format(time.RFC822)
}
