// Package livestream deals with purely live content, not pre-rec hence livestreaming
// and not streaming.
package livestream

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/jmoiron/sqlx/types"

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
		yt            *youtube.YouTube
	}
	// EditLivestream are parameters required to create or update a livestream.
	EditLivestream struct {
		Title          string    `json:"title" form:"title"`
		Description    string    `json:"description" form:"description"`
		ScheduledStart time.Time `json:"scheduledStart" form:"scheduledStart"`
		ScheduledEnd   time.Time `json:"scheduledEnd" form:"scheduledEnd"`
		Visibility     string    `json:"visbility" form:"visibility"`
		Thumbnail      string    `json:"thumbnail" form:"thumbnail"`
	}
	// Livestream is the metadata of a stream and the links to external
	// platforms.
	Livestream struct {
		ID             int       `db:"livestream_id" json:"livestreamID"`
		StreamKey      string    `db:"stream_key" json:"streamKey"`
		Status         string    `db:"status" json:"status"`
		Title          string    `db:"title" json:"title"`
		Description    string    `db:"description" json:"description"`
		ScheduledStart time.Time `db:"scheduled_start" json:"scheduledStart"`
		ScheduledEnd   time.Time `db:"scheduled_end" json:"scheduledEnd"`
		Visibility     string    `db:"visibility" json:"visbility"`
	}
	// ConsumeLivestream provides the links of a given stream key.
	ConsumeLivestream struct {
		ID        int    `db:"livestream_id" json:"livestreamID"`
		StreamKey string `db:"stream_key" json:"streamKey"`
	}
)

// New creates an instance of livestreamer.
func New(c Config, db *sqlx.DB, mcr *mcr.MCR, yt *youtube.YouTube) *Livestreamer {
	return &Livestreamer{
		ingestAddress: c.IngestAddress,
		db:            db,
		mcr:           mcr,
		yt:            yt,
	}
}

var (
	// ErrTitleEmpty when the title is empty.
	ErrTitleEmpty = errors.New("title is empty")
	// ErrTitleTooLong when the title is too long.
	ErrTitleTooLong = errors.New("title is too long, max 100 characters")
	// ErrDescriptionTooLong when the description is too long.
	ErrDescriptionTooLong = errors.New("description is too long, max 5000 characters")
	// ErrVisibilityInvalid when the given visibility option is invalid.
	ErrVisibilityInvalid = errors.New("invalid visibility option")
	// ErrStartAfterEnd when the livestream is scheduled to start after the end time.
	ErrStartAfterEnd = errors.New("scheduled start cannot be after the scheduled end")
	// ErrStartInPast when the start is in the past.
	ErrStartInPast = errors.New("start time cannot be in the past")
)

// New creates a livestream.
func (ls *Livestreamer) New(ctx context.Context, strm EditLivestream) (int, error) {
	if strm.Title == "" {
		return 0, ErrTitleEmpty
	}
	if len(strm.Title) > 100 {
		return 0, ErrTitleTooLong
	}
	if len(strm.Description) > 5000 {
		return 0, ErrDescriptionTooLong
	}
	if strm.Visibility != "public" && strm.Visibility != "unlisted" && strm.Visibility != "private" {
		return 0, ErrVisibilityInvalid
	}
	if !strm.ScheduledStart.Before(strm.ScheduledEnd) {
		return 0, ErrStartAfterEnd
	}
	if strm.ScheduledStart.Before(time.Now()) {
		return 0, ErrStartInPast
	}
	ingestKey := ls.generateStreamkey()
	strmID := 0
	err := ls.db.GetContext(ctx, &strmID, `
		INSERT INTO livestreams (
			stream_key,
			status,
			title,
			description,
			scheduled_start,
			scheduled_end,
			visibility
			) VALUES ($1, 'pending', $2, $3, $4, $5, $6)
			RETURNING livestream_id;`, ingestKey, strm.Title,
		strm.Description, strm.ScheduledStart, strm.ScheduledEnd, strm.Visibility)
	if err != nil {
		return 0, fmt.Errorf("failed to insert livestream: %w", err)
	}
	return strmID, nil
}

// Get a single livestream.
func (ls *Livestreamer) Get(ctx context.Context, livestreamID int) (Livestream, error) {
	strm := Livestream{}
	err := ls.db.GetContext(ctx, &strm, `
		SELECT
			livestream_id, stream_key, status, title, description, scheduled_start,
			scheduled_end, visibility
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
		SELECT livestream_id, stream_key, status, title 
		FROM livestreams;
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to get list of livestreams: %w", err)
	}
	return strms, nil
}

// Update a livestream.
func (ls *Livestreamer) Update(ctx context.Context, livestreamID int, strm EditLivestream) error {
	if strm.Title == "" {
		return ErrTitleEmpty
	}
	if len(strm.Title) > 100 {
		return ErrTitleTooLong
	}
	if len(strm.Description) > 5000 {
		return ErrDescriptionTooLong
	}
	if strm.Visibility != "public" && strm.Visibility != "unlisted" && strm.Visibility != "private" {
		return ErrVisibilityInvalid
	}
	if !strm.ScheduledStart.Before(strm.ScheduledEnd) {
		return ErrStartAfterEnd
	}
	if strm.ScheduledStart.Before(time.Now()) {
		return ErrStartInPast
	}

	_, err := ls.db.ExecContext(ctx, `
		UPDATE livestreams SET
			title = $1,
			description = $2,
			scheduled_start = $3,
			scheduled_end = $4,
			visibility = $5
		WHERE livestream_id = $6;`, strm.Title, strm.Description, strm.ScheduledStart,
		strm.ScheduledEnd, strm.Visibility, livestreamID)
	if err != nil {
		return fmt.Errorf("failed to update livestream: %w", err)
	}

	links, err := ls.ListLinks(ctx, livestreamID)
	if err != nil {
		return fmt.Errorf("failed to list links: %w", err)
	}

	for _, link := range links {
		switch link.IntegrationType {
		case LinkMCR:
			playoutID, err := strconv.Atoi(link.IntegrationID)
			if err != nil {
				return fmt.Errorf("failed to parse string to int: %w", err)
			}

			err = ls.mcr.UpdatePlayout(ctx, playoutID, mcr.EditPlayout{
				Title:          strm.Title,
				Description:    strm.Description,
				ScheduledStart: strm.ScheduledStart,
				ScheduledEnd:   strm.ScheduledEnd,
				Visibility:     strm.Visibility,
			})
			if err != nil {
				return fmt.Errorf("failed to update playout: %w", err)
			}

		case LinkYTNew:
			b, err := ls.yt.GetBroadcast(ctx, link.IntegrationID)
			if err != nil {
				return fmt.Errorf("failed to get broadcast: %w", err)
			}
			yt, err := ls.yt.GetYouTuber(b.AccountID)
			if err != nil {
				return fmt.Errorf("failed to get youtuber: %w", err)
			}
			err = yt.UpdateBroadcast(ctx, b.ID, youtube.EditBroadcast{
				Title:          strm.Title,
				Description:    strm.Description,
				ScheduledStart: strm.ScheduledStart,
				ScheduledEnd:   strm.ScheduledEnd,
				Visibility:     strm.Visibility,
			})
			if err != nil {
				return fmt.Errorf("failed to update yt-new broadcast: %w", err)
			}

		case LinkYTExisting: // nothing to do
		case LinkRTMPOutput: // nothing to do
		}
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

// Delete removes a livestream and it's associated links.
func (ls *Livestreamer) Delete(ctx context.Context, strm Livestream) error {
	links, err := ls.ListLinks(ctx, strm.ID)
	if err != nil {
		return fmt.Errorf("failed to list links: %w", err)
	}

	for _, link := range links {
		err = ls.DeleteLink(ctx, link)
		if err != nil {
			return fmt.Errorf("failed to delete link: %w", err)
		}
	}

	_, err = ls.db.ExecContext(ctx, `
		DELETE FROM livestreams
		WHERE livestream_id = $1;
	`, strm.ID)
	if err != nil {
		return fmt.Errorf("failed to delete livestream from store: %w", err)
	}

	return nil
}

func (ls *Livestreamer) ListEvents(ctx context.Context, strmID int) ([]Event, error) {
	// trying to directly unmarshal a JSONB field will result in it being base64 encoded
	// (see: https://github.com/jmoiron/sqlx/issues/133)
	var evts []struct {
		EventWithoutData
		Data types.JSONText `db:"event_data"`
	}
	err := ls.db.SelectContext(ctx, &evts, `
		SELECT livestream_event_id, event_type, event_data, event_time
		FROM livestream_events
		WHERE livestream_id = $1
		ORDER BY event_time ASC;
	`, strmID)
	if err != nil {
		return nil, fmt.Errorf("failed to get events: %w", err)
	}

	result := make([]Event, 0, len(evts))
	for _, evt := range evts {
		data, err := UnmarshalEventPayload(evt.Type, json.RawMessage(evt.Data))
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal event data: %w", err)
		}
		result = append(result, Event{
			EventWithoutData: evt.EventWithoutData,
			Data:             data,
		})
	}
	return result, nil
}

func (ls *Livestreamer) CreateEvent(ctx context.Context, strmID int, typ EventType, payload EventPayload) error {
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}
	_, err = ls.db.ExecContext(ctx, `
		INSERT INTO livestream_events (livestream_id, event_type, event_data)
		VALUES ($1, $2, $3::jsonb);
	`, strmID, typ, payloadJSON)
	if err != nil {
		return fmt.Errorf("failed to insert event: %w", err)
	}
	return nil
}

// PrettyDateTime formats dates to a more readable string.
func (strm Livestream) PrettyDateTime(ts time.Time) string {
	if ts.After(time.Now().Add(time.Hour * 24)) {
		return ts.Format("15:04 02/01")
	}
	return ts.Format("15:04")
}
