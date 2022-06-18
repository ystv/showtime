package youtube

import (
	"context"
	"fmt"

	"google.golang.org/api/youtube/v3"
)

type (
	// Stream is effectively what you know as the stream key.
	Stream struct {
		ID     string
		Title  string
		Status string
		Ingest Ingest
	}
	// Ingest is the config required to stream to YouTube's ingest.
	Ingest struct {
		Address string
		Key     string
	}
	// NewStream creates a new livestream.
	NewStream struct {
		Title       string
		Description string
		FrameRate   string
		IngestType  string
		Resolution  string
	}
)

func (y *YouTuber) switchStream(ctx context.Context, broadcastID string, streamID string) error {
	bind := y.yt.LiveBroadcasts.Bind(broadcastID, []string{"id"})
	if streamID != "" {
		bind = bind.StreamId(streamID)
	}
	_, err := bind.Do()
	if err != nil {
		return fmt.Errorf("failed to bind stream: %w", err)
	}
	return nil
}

func (y *YouTuber) newStream(ctx context.Context) (Stream, error) {
	newStream := NewStream{
		Title:       "YSTV Media Services",
		Description: "Auto-generated stream key made by YSTV ShowTime!",
		FrameRate:   "variable",
		IngestType:  "rtmp",
		Resolution:  "variable",
	}

	req := &youtube.LiveStream{
		Snippet: &youtube.LiveStreamSnippet{
			Title:       newStream.Title,
			Description: newStream.Description,
		},
		Cdn: &youtube.CdnSettings{
			IngestionType: newStream.IngestType,
			FrameRate:     newStream.FrameRate,
			Resolution:    newStream.Resolution,
		},
	}

	ytStream, err := y.yt.LiveStreams.Insert([]string{"id,snippet,cdn"}, req).Do()
	if err != nil {
		return Stream{}, fmt.Errorf("failed to create stream: %w", err)
	}

	stream := Stream{
		ID:     ytStream.Id,
		Title:  ytStream.Snippet.Title,
		Status: "noData",
		Ingest: Ingest{
			Address: ytStream.Cdn.IngestionInfo.IngestionAddress,
			Key:     ytStream.Cdn.IngestionInfo.StreamName,
		},
	}

	return stream, nil
}
