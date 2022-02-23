package youtube

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/ystv/showtime/auth"
	"google.golang.org/api/youtube/v3"
)

type (
	YouTuber struct {
		yt *youtube.Service
		db *sqlx.DB
	}
	Broadcast struct {
		ID        string `json:"id"`
		Title     string `json:"title"`
		StartTime string `json:"startTime"`
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

var Schema = `
CREATE TABLE youtube_broadcasts (
	broadcast_id text NOT NULL PRIMARY KEY,
	ingest_address text NOT NULL,
	stream_name text NOT NULL
);
`

func New(db *sqlx.DB, auth *auth.Auther) (*YouTuber, error) {
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
		db: db,
	}, nil
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

	_, err = y.db.ExecContext(ctx, `
		INSERT INTO youtube_broadcasts (
			broadcast_id,
			ingest_address,
			stream_name
		) VALUES ($1, $2, $3);`, broadcastID, stream.Ingest.Address, stream.Ingest.Name)

	if err != nil {
		return fmt.Errorf("failed to add broadcast to db: %w", err)
	}

	return nil
}

func (y *YouTuber) DisableShowTimeForBroadcast(ctx context.Context, broadcastID string) error {
	err := y.switchStream(ctx, broadcastID, "")
	if err != nil {
		return fmt.Errorf("failed to switch stream: %w", err)
	}

	_, err = y.db.ExecContext(ctx, `
	DELETE FROM youtube_broadcasts WHERE broadcast_id = $1;`, broadcastID)
	if err != nil {
		return fmt.Errorf("failed to delete broadcast from db: %w", err)
	}

	return nil
}

func (b *Broadcast) PrettyDateTime() string {
	ts, err := time.Parse(time.RFC3339, b.StartTime)
	if err != nil {
		return err.Error()
	}

	return ts.Format(time.RFC822)
}
