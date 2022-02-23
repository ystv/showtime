package playout

import (
	"context"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/tjarratt/babble"
)

type (
	Playouter struct {
		db *sqlx.DB
	}
	NewPlayout struct {
		Title string `db:"title" json:"title"`
	}
	Playout struct {
		PlayoutID     int    `db:"playout_id" json:"playoutID"`
		Title         string `db:"title" json:"title"`
		WebsiteLinkID string `db:"website_link_id" json:"websiteLinkID"`
		YouTubeLinkID string `db:"youtube_link_id" json:"youtubeLinkID"`
	}
	ConsumePlayout struct {
		Title     string `db:"title" json:"title"`
		StreamKey string `db:"stream_key" json:"streamKey"`
	}
)

var Schema = `
CREATE TABLE playouts(
	playout_id integer NOT NULL PRIMARY KEY AUTOINCREMENT,
	title text NOT NULL,
	stream_key text NOT NULL,
	website_link_id text NOT NULL,
	youtube_link_id text NOT NULL
);
`

func New(db *sqlx.DB) *Playouter {
	return &Playouter{db: db}
}

func (p *Playouter) New(ctx context.Context, po NewPlayout) error {
	streamKey := p.generateStreamkey()
	_, err := p.db.ExecContext(ctx, `
		INSERT INTO playouts (
			title,
			stream_key,
			website_link_id,
			youtube_link_id
			) VALUES ($1, $2, '', '');`, po.Title, streamKey)
	if err != nil {
		return fmt.Errorf("failed to insert playout: %w", err)
	}
	return nil
}

func (p *Playouter) List(ctx context.Context) ([]Playout, error) {
	po := []Playout{}
	err := p.db.SelectContext(ctx, &po, `
		SELECT playout_id, title, website_link_id, youtube_link_id
		FROM playouts;	
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to get list of playouts")
	}
	return po, nil
}

func (p *Playouter) Update(ctx context.Context, po Playout) error {
	_, err := p.db.ExecContext(ctx, `UPDATE playouts SET
			title = $1,
			website_link_id = $2,
			youtube_link_id = $3
		WHERE playout_id = $4;`, po.Title, po.WebsiteLinkID, po.YouTubeLinkID, po.PlayoutID)
	if err != nil {
		return fmt.Errorf("failed to update playout: %w", err)
	}
	return nil
}

func (p *Playouter) generateStreamkey() string {
	babbler := babble.NewBabbler()
	babbler.Separator = "-"
	babbler.Count = 3
	return strings.ToLower(babbler.Babble())
}
