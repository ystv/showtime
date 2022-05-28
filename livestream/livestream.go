// Package livestream deals with purely live content, not pre-rec hence livestreaming
// and not streaming.
package livestream

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/ystv/showtime/mcr"
	"github.com/ystv/showtime/youtube"
)

type (
	// Config configures livestreamer.
	Config struct {
		IngestAddress string
	}
	// Livestreamer lets links be created to livestreaming platforms.
	Livestreamer struct {
		ingestAddress string
		db            *sqlx.DB
		mcr           *mcr.MCR
		yt            *youtube.YouTuber
	}
	// NewLivestream creates a new livestream.
	NewLivestream struct {
		Title string `json:"title" form:"title"`
	}
	// Livestream is the metadata of a stream and the links to external
	// platforms.
	Livestream struct {
		LivestreamID  int    `db:"livestream_id" json:"livestreamID"`
		StreamKey     string `db:"stream_key" json:"streamKey"`
		Status        string `db:"status" json:"status"`
		Title         string `db:"title" json:"title"`
		MCRLinkID     string `db:"mcr_link_id" json:"mcrLinkID"`         // MCR's playoutID
		YouTubeLinkID string `db:"youtube_link_id" json:"youtubeLinkID"` // YT's broadcastID
	}
	// ConsumeLivestream provides the links of a given stream key.
	ConsumeLivestream struct {
		StreamKey     string `db:"stream_key" json:"streamKey"`
		MCRLinkID     string `db:"mcr_link_id" json:"mcrLinkID"`
		YouTubeLinkID string `db:"youtube_link_id" json:"youtubeLinkID"`
	}
)

// Schema represents the livestream package in the database.
var Schema = `
CREATE TABLE livestreams(
	livestream_id integer NOT NULL PRIMARY KEY AUTOINCREMENT,
	status text NOT NULL,
	stream_key text NOT NULL UNIQUE,
	title text NOT NULL,
	mcr_link_id text NOT NULL,
	youtube_link_id text NOT NULL
);
`

// New creates an instance of livestreamer.
func New(c Config, db *sqlx.DB, mcr *mcr.MCR, yt *youtube.YouTuber) *Livestreamer {
	return &Livestreamer{
		ingestAddress: c.IngestAddress,
		db:            db,
		mcr:           mcr,
		yt:            yt,
	}
}

// New creates a livestream.
func (ls *Livestreamer) New(ctx context.Context, strm NewLivestream) error {
	streamKey := ls.generateStreamkey()
	_, err := ls.db.ExecContext(ctx, `
		INSERT INTO livestreams (
			stream_key,
			status,
			title,
			mcr_link_id,
			youtube_link_id
			) VALUES ($1, 'pending', $2, '', '');`, streamKey, strm.Title)
	if err != nil {
		return fmt.Errorf("failed to insert livestream: %w", err)
	}
	return nil
}

// Get a single livestream.
func (ls *Livestreamer) Get(ctx context.Context, livestreamID int) (Livestream, error) {
	strm := Livestream{}
	err := ls.db.GetContext(ctx, &strm, `
		SELECT livestream_id, stream_key, status, title, mcr_link_id, youtube_link_id
		FROM livestreams
		WHERE livestream_id  = $1;
	`, livestreamID)
	if err != nil {
		return Livestream{}, fmt.Errorf("failed to get livestream: %w", err)
	}
	return strm, nil
}

// List all livestreams.
func (ls *Livestreamer) List(ctx context.Context) ([]Livestream, error) {
	strms := []Livestream{}
	err := ls.db.SelectContext(ctx, &strms, `
		SELECT livestream_id, stream_key, status, title, mcr_link_id, youtube_link_id
		FROM livestreams;
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to get list of livestreams: %w", err)
	}
	return strms, nil
}

// Update a livestream.
func (ls *Livestreamer) Update(ctx context.Context, strm Livestream) error {
	_, err := ls.db.ExecContext(ctx, `
		UPDATE livestreams SET
			title = $1,
			website_link_id = $2,
			youtube_link_id = $3
		WHERE playout_id = $4;`, strm.Title, strm.MCRLinkID, strm.YouTubeLinkID, strm.LivestreamID)
	if err != nil {
		return fmt.Errorf("failed to update playout: %w", err)
	}
	return nil
}

// UpdateMCRLink updates only the MCR playout link on a livestream.
func (ls *Livestreamer) UpdateMCRLink(ctx context.Context, livestreamID int, linkID string) error {
	_, err := ls.db.ExecContext(ctx, `
	UPDATE livestreams SET
		mcr_link_id = $1
	WHERE livestream_id = $2;`, linkID, livestreamID)
	if err != nil {
		return fmt.Errorf("failed to update mcr link id: %w", err)
	}
	return nil
}

// UpdateYouTubeLink updates only the YouTube link on a livestream.
func (ls *Livestreamer) UpdateYouTubeLink(ctx context.Context, livestreamID int, linkID string) error {
	_, err := ls.db.ExecContext(ctx, `
	UPDATE livestreams SET
		youtube_link_id = $1
	WHERE livestream_id = $2;`, linkID, livestreamID)
	if err != nil {
		return fmt.Errorf("failed to update youtube link id: %w", err)
	}
	return nil
}

// updateStatus updates only the status on a livestream.
func (ls *Livestreamer) updateStatus(ctx context.Context, livestreamID int, status string) error {
	_, err := ls.db.ExecContext(ctx, `
		UPDATE livestreams SET
			status = $1
		WHERE livestream_id = $2`, status, livestreamID)
	if err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}
	return nil
}
