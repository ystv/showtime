package livestream

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"

	"github.com/ystv/showtime/mcr"
)

type (
	// Link is a relationship between a livestream and an integration.
	Link struct {
		ID              int             `db:"link_id"`
		IntegrationType IntegrationType `db:"integration_type"`
		IntegrationID   string          `db:"integration_id"`
	}
	// NewLinkParams are params to create a new link.
	NewLinkParams struct {
		LivestreamID    int
		IntegrationType IntegrationType
		IntegrationID   string
	}
	// IntegrationType is a type of intergration with a platform.
	IntegrationType string
)

const (
	// LinkMCR enables full integration with MCR.
	LinkMCR IntegrationType = "mcr"
	// LinkYTNew enables full integration with YouTube, creating a new YouTube broadcast.
	LinkYTNew IntegrationType = "yt-new"
	// LinkYTExisting enables partial integration with an existing YouTube broadcast.
	LinkYTExisting IntegrationType = "yt-existing"
	// LinkRTMPOutput enables partial integration to an RTMP endpoint.
	LinkRTMPOutput IntegrationType = "rtmp"
)

var (
	// ErrUnkownIntegrationType when the integration type is unknown.
	ErrUnkownIntegrationType = errors.New("unknown integration type")
)

func (i IntegrationType) String() string {
	return string(i)
}

// NewLink creates a new relationship between a
func (ls *Livestreamer) NewLink(ctx context.Context, l NewLinkParams) (Link, error) {
	linkID := 0
	err := ls.db.GetContext(ctx, &linkID, `
		INSERT INTO links (livestream_id, integration_type, integration_id)
		VALUES ($1, $2, $3)
		RETURNING link_id;
	`, l.LivestreamID, l.IntegrationType, l.IntegrationID)
	if err != nil {
		return Link{}, fmt.Errorf("failed to create link: %w", err)
	}
	if err := ls.CreateEvent(ctx, l.LivestreamID, EventLinked, EventLinkedPayload{
		IntegrationType: l.IntegrationType,
		IntegrationID:   l.IntegrationID,
	}); err != nil {
		log.Printf("failed to log link event: %v", err)
	}
	return Link{ID: linkID, IntegrationType: l.IntegrationType, IntegrationID: l.IntegrationID}, nil
}

// GetLink returns a single link.
func (ls *Livestreamer) GetLink(ctx context.Context, linkID int) (Link, error) {
	link := Link{}
	err := ls.db.GetContext(ctx, &link, `
		SELECT link_id, integration_type, integration_id
		FROM links
		WHERE link_id = $1;
	`, linkID)
	return link, err
}

// ListLinks returns a list of links for a given livestream.
func (ls *Livestreamer) ListLinks(ctx context.Context, livestreamID int) ([]Link, error) {
	links := []Link{}
	err := ls.db.SelectContext(ctx, &links, `
		SELECT link_id, integration_type, integration_id
		FROM links
		WHERE livestream_id = $1;
	`, livestreamID)
	return links, err
}

// DeleteLink removes a relationship between a livestream and an integration.
func (ls *Livestreamer) DeleteLink(ctx context.Context, link Link) error {
	var streamID int
	err := ls.db.GetContext(ctx, &streamID, `SELECT livestream_id FROM links WHERE link_id = $1;`, link.ID)
	if err != nil {
		return fmt.Errorf("failed to get livestream id: %w", err)
	}
	switch link.IntegrationType {
	case LinkMCR:
		playoutID, err := strconv.Atoi(link.IntegrationID)
		if err != nil {
			return fmt.Errorf("failed to convert integration id to playout id: %w", err)
		}

		err = ls.mcr.DeletePlayout(ctx, playoutID)
		if err != nil {
			if !errors.Is(err, mcr.ErrPlayoutNotFound) {
				return fmt.Errorf("failed to delete playout: %w", err)
			}
		}

	case LinkYTNew:
		err := ls.yt.DeleteBroadcast(ctx, link.IntegrationID)
		if err != nil {
			return fmt.Errorf("failed to delete broadcast: %w", err)
		}

	case LinkYTExisting:
		err := ls.yt.DeleteExistingBroadcast(ctx, link.IntegrationID)
		if err != nil {
			return fmt.Errorf("failed to delete existing broadcast: %w", err)
		}

	case LinkRTMPOutput:
		rtmpOutputID, err := strconv.Atoi(link.IntegrationID)
		if err != nil {
			return fmt.Errorf("failed to convert integration id to rtmp output id: %w", err)
		}
		err = ls.DeleteRTMPOutput(ctx, rtmpOutputID)
		if err != nil {
			return fmt.Errorf("failed to delete rtmp output: %w", err)
		}

	default:
		return ErrUnkownIntegrationType
	}
	_, err = ls.db.ExecContext(ctx, `
		DELETE FROM links
		WHERE link_id = $1;
	`, link.ID)
	if err != nil {
		return fmt.Errorf("failed to delete link from store: %w", err)
	}
	if err := ls.CreateEvent(ctx, streamID, EventUnlinked, EventUnlinkedPayload{
		IntegrationType: link.IntegrationType,
		IntegrationID:   link.IntegrationID,
	}); err != nil {
		log.Printf("failed to log unlink event: %v", err)
	}
	return nil
}
