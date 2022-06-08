package youtube

import (
	"context"
	"errors"
	"fmt"
	"time"

	"google.golang.org/api/youtube/v3"
)

type (
	// Broadcast are the livestreams which are watchable as videos.
	Broadcast struct {
		ID             string `db:"broadcast_id" json:"id"`
		AccountID      int    `db:"account_id" json:"accountID"`
		IngestAddress  string `db:"ingest_address" json:"ingestAddress"`
		IngestKey      string `db:"ingest_key" json:"ingestKey"`
		Title          string `db:"title" json:"title"`
		Description    string `db:"description" json:"description"`
		ScheduledStart string `db:"scheduled_start" json:"scheduledStart"`
		ScheduledEnd   string `db:"scheduled_end" json:"scheduledEnd"`
		Visibility     string `db:"visibility" json:"visibility"`
	}

	// EditBroadcast are parameters required to create or update a broadcast.
	EditBroadcast struct {
		Title          string
		Description    string
		ScheduledStart time.Time
		ScheduledEnd   time.Time
		Visibility     string
	}
)

var (
	// ErrBroadcastNotFound when the broadcast can't be found.
	ErrBroadcastNotFound = errors.New("broadcast not found")
)

// StartBroadcast triggers the state to be updated to "live" making it
// visible to the audience.
func (y *YouTuber) StartBroadcast(ctx context.Context, b Broadcast) error {
	_, err := y.yt.LiveBroadcasts.Transition("live", b.ID, []string{}).Do()
	if err != nil {
		return fmt.Errorf("failed to transition broadcast to live state: %w", err)
	}
	return nil
}

// EndBroadcast triggers the state to be updated to "complete" making it
// marked as done.
func (y *YouTuber) EndBroadcast(ctx context.Context, b Broadcast) error {
	_, err := y.yt.LiveBroadcasts.Transition("complete", b.ID, []string{}).Do()
	if err != nil {
		return fmt.Errorf("failed to transition broadcast to complete state: %w", err)
	}
	return nil
}

// NewBroadcast creates a new broadcast.
func (y *YouTuber) NewBroadcast(ctx context.Context, p EditBroadcast) (Broadcast, error) {
	ytBroadcast, err := y.yt.LiveBroadcasts.Insert([]string{"snippet", "status"}, &youtube.LiveBroadcast{
		Snippet: &youtube.LiveBroadcastSnippet{
			Title:              p.Title,
			Description:        p.Description,
			ScheduledStartTime: p.ScheduledStart.Format(time.RFC3339),
			ScheduledEndTime:   p.ScheduledEnd.Format(time.RFC3339),
		},
		Status: &youtube.LiveBroadcastStatus{
			PrivacyStatus:           p.Visibility,
			SelfDeclaredMadeForKids: false,
		},
	}).Do()
	if err != nil {
		return Broadcast{}, fmt.Errorf("failed to insert broadcast: %w", err)
	}

	stream, err := y.newStream(ctx)
	if err != nil {
		return Broadcast{}, fmt.Errorf("failed to create stream: %w", err)
	}

	err = y.switchStream(ctx, ytBroadcast.Id, stream.ID)
	if err != nil {
		return Broadcast{}, fmt.Errorf("failed to switch stream: %w", err)
	}

	_, err = y.db.ExecContext(ctx, `
		INSERT INTO youtube.broadcasts (
			broadcast_id,	account_id,	ingest_address, ingest_key, title, description,
			scheduled_start, scheduled_end, visibility
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9);`, ytBroadcast.Id,
		y.accountID, stream.Ingest.Address, stream.Ingest.Key, p.Title,
		p.Description, p.ScheduledStart, p.ScheduledEnd, p.Visibility)
	if err != nil {
		return Broadcast{}, fmt.Errorf("failed to add broadcast to store: %w", err)
	}

	b := Broadcast{
		ID:             ytBroadcast.Id,
		AccountID:      y.accountID,
		IngestAddress:  stream.Ingest.Address,
		IngestKey:      stream.Ingest.Key,
		Title:          p.Title,
		Description:    p.Description,
		ScheduledStart: p.ScheduledStart.Format(time.RFC3339),
		ScheduledEnd:   p.ScheduledEnd.Format(time.RFC3339),
		Visibility:     p.Visibility,
	}

	return b, nil
}

// UpdateBroadcast updates an existing YouTube broadcast.
func (y *YouTuber) UpdateBroadcast(ctx context.Context, broadcastID string, p EditBroadcast) error {
	_, err := y.yt.LiveBroadcasts.Update([]string{"snippet", "status"}, &youtube.LiveBroadcast{
		Id: broadcastID,
		Snippet: &youtube.LiveBroadcastSnippet{
			Title:              p.Title,
			Description:        p.Description,
			ScheduledStartTime: p.ScheduledStart.Format(time.RFC3339),
			ScheduledEndTime:   p.ScheduledEnd.Format(time.RFC3339),
		},
		Status: &youtube.LiveBroadcastStatus{
			PrivacyStatus: p.Visibility,
		},
	}).Do()
	if err != nil {
		return fmt.Errorf("failed to update broadcast in yt: %w", err)
	}

	_, err = y.db.ExecContext(ctx, `
		UPDATE youtube.broadcasts SET
			title = $1,
			description = $2,
			scheduled_start = $3,
			scheduled_end = $4,
			visibility = $5
		WHERE broadcast_id = $6;`, p.Title, p.Description, p.ScheduledStart,
		p.ScheduledEnd, p.Visibility, broadcastID)
	if err != nil {
		return fmt.Errorf("failed to update broadcast in store: %w", err)
	}

	return nil
}

// DeleteBroadcast disables ShowTime! integration and deletes the broadcast on
// YouTube.
func (y *YouTube) DeleteBroadcast(ctx context.Context, broadcastID string) error {
	b, err := y.GetBroadcast(ctx, broadcastID)
	if err != nil {
		return fmt.Errorf("failed to get broadcast: %w", err)
	}
	yt, err := y.GetYouTuber(b.AccountID)
	if err != nil {
		return fmt.Errorf("failed to get youtuber: %w", err)
	}
	return yt.DeleteBroadcast(ctx, b)
}

// DeleteBroadcast disables ShowTime! integration and deletes the
// broadcast on YouTube.
func (y *YouTuber) DeleteBroadcast(ctx context.Context, b Broadcast) error {
	err := y.switchStream(ctx, b.ID, "")
	if err != nil {
		return fmt.Errorf("failed to switch stream: %w", err)
	}

	err = y.yt.LiveBroadcasts.Delete(b.ID).Do()
	if err != nil {
		return fmt.Errorf("failed to delete broadcast in yt")
	}

	_, err = y.db.ExecContext(ctx, `
		DELETE FROM youtube.broadcasts
		WHERE broadcast_id = $1;`, b.ID)
	if err != nil {
		return fmt.Errorf("failed to delete broadcast in store: %w", err)
	}

	return nil
}

// NewExistingBroadcast enables ShowTime! integration on an existing broadcast.
//
// The incoming video stream will be forwarded to YouTube and the state can be
// controlled.
func (y *YouTuber) NewExistingBroadcast(ctx context.Context, broadcastID string) error {
	b, err := y.getBroadcastDirect(ctx, broadcastID)
	if err != nil {
		return fmt.Errorf("failed to get broadcast directly: %w", err)
	}

	stream, err := y.newStream(ctx)
	if err != nil {
		return fmt.Errorf("failed to create stream: %w", err)
	}

	err = y.switchStream(ctx, broadcastID, stream.ID)
	if err != nil {
		return fmt.Errorf("failed to switch stream: %w", err)
	}

	_, err = y.db.ExecContext(ctx, `
		INSERT INTO youtube.broadcasts (
			broadcast_id,	account_id,	ingest_address, ingest_key,	title, description,
			scheduled_start, scheduled_end,	visibility
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9);`,
		broadcastID, y.accountID, stream.Ingest.Address, stream.Ingest.Key,
		b.Title, b.Description, b.ScheduledStart, b.ScheduledEnd, b.Visibility)
	if err != nil {
		return fmt.Errorf("failed to add broadcast to store: %w", err)
	}

	return nil
}

// DeleteExistingBroadcast disables ShowTime! integration and leaves the
// broadcast to still exist on YouTube.
func (y *YouTube) DeleteExistingBroadcast(ctx context.Context, broadcastID string) error {
	b, err := y.GetBroadcast(ctx, broadcastID)
	if err != nil {
		return fmt.Errorf("failed to get broadcast details: %w", err)
	}
	yt, err := y.GetYouTuber(b.AccountID)
	if err != nil {
		return fmt.Errorf("failed to get youtuber: %w", err)
	}
	return yt.DeleteExistingBroadcast(ctx, b)
}

// DeleteExistingBroadcast disables ShowTime! integration and leaves the
// broadcast to still exist on YouTube.
//
// Deletes the generated stream key to the default and updates the database to
// de-link the broadcast.
func (y *YouTuber) DeleteExistingBroadcast(ctx context.Context, b Broadcast) error {
	err := y.switchStream(ctx, b.ID, "")
	if err != nil {
		return fmt.Errorf("failed to switch stream: %w", err)
	}

	_, err = y.db.ExecContext(ctx, `
		DELETE FROM youtube.broadcasts
		WHERE broadcast_id = $1;`, b.ID)
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
	req := y.yt.LiveBroadcasts.List([]string{"id", "snippet", "status"})
	req.BroadcastStatus("upcoming")
	ytBroadcasts, err := req.Do()
	if err != nil {
		return nil, fmt.Errorf("failed to list broadcasts: %w", err)
	}
	for _, ytBroadcast := range ytBroadcasts.Items {
		broadcast := Broadcast{
			ID:             ytBroadcast.Id,
			Title:          ytBroadcast.Snippet.Title,
			Description:    ytBroadcast.Snippet.Description,
			ScheduledStart: ytBroadcast.Snippet.ScheduledStartTime,
			ScheduledEnd:   ytBroadcast.Snippet.ScheduledEndTime,
			Visibility:     ytBroadcast.Status.PrivacyStatus,
		}
		broadcasts = append(broadcasts, broadcast)
	}
	return broadcasts, nil
}

// GetBroadcast retrives stream ingest and name details.
//
// Retrieved from store which is quick.
func (y *YouTube) GetBroadcast(ctx context.Context, broadcastID string) (Broadcast, error) {
	b := Broadcast{}
	err := y.db.GetContext(ctx, &b, `
		SELECT broadcast_id, account_id, ingest_address, ingest_key, title,
					 description, scheduled_start, scheduled_end, visibility
		FROM youtube.broadcasts
		WHERE broadcast_id = $1;
	`, broadcastID)
	return b, err
}

// GetBroadcast retrives stream ingest and name details.
//
// Retrieved from store which is quick.
func (y *YouTuber) GetBroadcast(ctx context.Context, broadcastID string) (Broadcast, error) {
	b := Broadcast{}
	err := y.db.GetContext(ctx, &b, `
		SELECT broadcast_id, account_id, ingest_address, ingest_key, title,
					 description, scheduled_start, scheduled_end, visbility
		FROM youtube.broadcasts
		WHERE broadcast_id = $1;
	`, broadcastID)
	return b, err
}

// getBroadcastDirect gets a broadcast directly from YouTube.
func (y *YouTuber) getBroadcastDirect(ctx context.Context, broadcastID string) (Broadcast, error) {
	broadcasts, err := y.ListBroadcasts(ctx)
	if err != nil {
		return Broadcast{}, fmt.Errorf("failed to list broadcasts: %w", err)
	}
	for _, broadcast := range broadcasts {
		if broadcast.ID == broadcastID {
			return broadcast, nil
		}
	}
	return Broadcast{}, ErrBroadcastNotFound
}

// GetTotalLinkedBroadcasts returns the total count of linked broadcasts.
func (y *YouTuber) GetTotalLinkedBroadcasts(ctx context.Context) (int, error) {
	total := 0
	err := y.db.GetContext(ctx, &total, `
		SELECT COUNT(*)
		FROM youtube.broadcasts
		WHERE account_id = $1;
	`, y.accountID)
	if err != nil {
		return 0, fmt.Errorf("failed to get total linked broadcasts: %w", err)
	}
	return total, nil
}

// ListShowTimedBroadcasts returns a list of broadcasts that have ShowTime enabled.
func (y *YouTuber) ListShowTimedBroadcasts(ctx context.Context) ([]Broadcast, error) {
	broadcasts := []Broadcast{}
	err := y.db.SelectContext(ctx, &broadcasts, `
		SELECT broadcast_id
		FROM youtube.broadcasts
		WHERE account_id = $1;
	`, y.accountID)
	return broadcasts, err
}

// PrettyDateTime formats dates to a more readable string.
func (b *Broadcast) PrettyDateTime() string {
	ts, err := time.Parse(time.RFC3339, b.ScheduledStart)
	if err != nil {
		return "start time not set"
	}

	return ts.Format(time.RFC822)
}
