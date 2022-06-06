package livestream

import (
	"context"
	"errors"
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
	// MCR enables full integration with MCR.
	MCR IntegrationType = "mcr"
	// YTExisting enables partial integration with an existing YouTube broadcast.
	YTExisting IntegrationType = "yt-existing"
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
	return Link{ID: linkID, IntegrationType: l.IntegrationType, IntegrationID: l.IntegrationID}, err
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
func (ls *Livestreamer) DeleteLink(ctx context.Context, linkID int) error {
	_, err := ls.db.ExecContext(ctx, `
		DELETE FROM links
		WHERE link_id = $1;
	`, linkID)
	return err
}