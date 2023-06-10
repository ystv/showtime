package mixer

import (
	"context"

	"github.com/jmoiron/sqlx"
)

type (
	Mixerer struct {
		db *sqlx.DB
	}
	// EditMixer represents the changable fields of a mixer.
	EditMixer struct {
		Address  string
		Username string
		Password string
		Type     MixerType
	}
	// Mixer represents a mixer.
	Mixer struct {
		ID       int       `db:"mixer_id"`
		Address  string    `db:"address"`
		Username string    `db:"username"`
		Password string    `db:"password"`
		Type     MixerType `db:"type"`
	}

	// MixerType represents a mixer type.
	MixerType string
)

const (
	// MixerBrave represents a Brave mixer.
	MixerBrave MixerType = "brave"
	// MixerOBS represents an OBS mixer.
	MixerOBS MixerType = "obs"
)

// New creates a new mixer instance.
func New(db *sqlx.DB) *Mixerer {
	return &Mixerer{db: db}
}

// New creates a new mixer.
func (m *Mixerer) New(ctx context.Context, edit EditMixer) (int, error) {
	mixerID := 0
	err := m.db.GetContext(ctx, &mixerID, `
		INSERT INTO mixers (address, username, password, type)
		VALUES ($1, $2, $3, $4)
		RETURNING mixer_id;
	`, edit.Address, edit.Username, edit.Password, edit.Type)
	return mixerID, err
}

// Get gets the mixer.
func (m *Mixerer) Get(ctx context.Context, mixerID int) (Mixer, error) {
	mixer := Mixer{}
	err := m.db.GetContext(ctx, &mixer, `
		SELECT mixer_id, address, username, password, type
		FROM mixers
		WHERE mixer_id = $1;
	`, mixerID)
	return mixer, err
}

// Update updates the mixer.
func (m *Mixerer) Update(ctx context.Context, mixerID int, edit EditMixer) error {
	_, err := m.db.ExecContext(ctx, `
		UPDATE mixers
		SET address = $1, username = $2, password = $3, type = $4
		WHERE mixer_id = $5;
	`, edit.Address, edit.Username, edit.Password, edit.Type, mixerID)
	return err
}

// Delete deletes the mixer.
func (m *Mixerer) Delete(ctx context.Context, mixerID int) error {
	_, err := m.db.ExecContext(ctx, `
		DELETE FROM mixers
		WHERE mixer_id = $1;
	`, mixerID)
	return err
}

// List lists all mixers excluding passwords.
func (m *Mixerer) List(ctx context.Context) ([]Mixer, error) {
	mixers := []Mixer{}
	err := m.db.SelectContext(ctx, &mixers, `
		SELECT mixer_id, address, username, type
		FROM mixers;
	`)
	return mixers, err
}
