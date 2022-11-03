package engine

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/ystv/showtime/brave"
)

type (
	// Enginer represents a collection of engines.
	Enginer struct {
		db    *sqlx.DB
		hosts map[int]*host
	}

	// Engine is a video pipeline.
	Engine struct {
		ID   int
		host *host
	}

	// host hosts engines.
	host struct {
		id      int    `db:"host_id"`
		address string `db:"address"`

		brave *brave.Braver
		db    *sqlx.DB
	}

	// Config stores application config
	Config struct {
		HostAddresses []string
	}
)

// Schema represents the engine package in the database.
var Schema = `
CREATE SCHEMA engine;

CREATE TABLE engine.hosts (
	host_id bigint GENERATED AS ALWAYS AS IDENTITY,
	address text NOT NULL UNIQUE,
	PRIMARY KEY host_id
);

CREATE TABLE engine.inputs (
	input_id bigint GENERATED AS ALWAYS AS IDENTITY
);

CREATE TABLE engine.outputs (
	output_id bigint GENERATED AS ALWAYS AS IDENTITY,
	mixer_id bigint NOT NULL,
	type text NOT NULL,
	address text NOT NULL,
	CONSTRAINT fk_mixer_id FOREIGN KEY(mixer_id) REFERENCES engine.mixers(mixer_id) ON DELETE CASCADE
);

CREATE TABLE engine.mixers (
	mixer_id bigint GENERTED AS ALWAYS AS IDENTITY,
	width int NOT NULL,
	height int NOT NULL
);

CREATE TABLE engine.engines (
	channel_id bigint NOT NULL,
	mixer_id integer NOT NULL,
	program_input_id bigint NOT NULL,
	continuity_input_id bigint NOT NULL,
	program_output_id bigint NOT NULL,
	CONSTRAINT fk_channel_id FOREIGN KEY(channel_id) REFERENCES mcr.channels(channel_id) ON DELETE CASCADE,
	CONSTRAINT fk_program_input_id FOREIGN KEY(program_input_id) REFERENCES engine.inputs(input_id),
	CONSTRAINT fk_continuity_input_id FOREIGN KEY(continuity_input_id) REFERENCES engine.inputs(input_id),
	PRIMARY KEY (channel_id)
);

CREATE TABLE engine.allocations (
	mixer_id int NOT NULL,
	host_id int NOT NULL,
	CONSTRAINT fk_mixer_id FOREIGN KEY(mixer_id) REFERENCES engine.mixers(mixer_id) ON DELETE CASCADE,
	CONSTRAINT fk_host_id FOREIGN KEY(host_id) REFERENCES engine.hosts(host_id) ON DELETE CASCADE,
	PRIMARY KEY (mixer_id, host_id)
);

CREATE TABLE engine.mappings (
	showtime_id bigint NOT NULL,
	type text NOT NULL,
	host_id bigint NOT NULL,
	mixer_id bigint NOT NULL,
	brave_id bigint NOT NULL,
	CONSTRAINT fk_host_id FOREIGN KEY(host_id) REFERENCES engine.hosts(host_id) ON DELETE CASCADE,
	CONSTRAINT fk_mixer_id FOREIGN KEY(mixer_id) REFERENCES engine.mixers(mixer_id) ON DELETE CASCADE,
	PRIMARY KEY (showtime_id, type)
);
`

// New creates a enginer instance.
func New(ctx context.Context, db *sqlx.DB, conf *Config) (*Enginer, error) {
	eng := &Enginer{
		db:    db,
		hosts: map[int]*host{},
	}

	configHosts := []host{}
	for _, address := range conf.HostAddresses {
		configHosts = append(configHosts, host{
			address: address,
		})
	}

	store, err := eng.listHostsFromStore(ctx)

	// Remove hosts no longer in config.
	for _, storeHost := range store {
		exists := false
		for _, configHost := range configHosts {
			if configHost.address == storeHost.address {
				exists = true
			}
		}
		if !exists {
			err = eng.deleteHost(ctx, storeHost)
			if err != nil {
				return nil, fmt.Errorf("failed to delete host: %w", err)
			}
		}
	}

	// Add new hosts that are not in the store.
	for _, configHost := range configHosts {
		exists := false
		for _, storeHost := range store {
			if configHost.address == storeHost.address {
				exists = true
			}
		}
		if !exists {
			err = eng.newHost(ctx, configHost.address)
			if err != nil {
				return nil, fmt.Errorf("failed to create host: %w", err)
			}
		}
	}

	return eng, nil
}

func (eng *Enginer) newHost(ctx context.Context, address string) error {
	h := &host{
		address: address,
		db:      eng.db,
	}
	err := eng.db.GetContext(ctx, h.id, `
		INSERT INTO engine.hosts (address)
		VALUES ($1)
		RETURNING host_id;`, address)
	if err != nil {
		return fmt.Errorf("failed to insert host: %w", err)
	}

	eng.hosts[h.id] = h

	return nil
}

func (eng *Enginer) listHostsFromStore(ctx context.Context) ([]host, error) {
	h := []host{}
	err := eng.db.SelectContext(ctx, &h, `
		SELECT host_id, address
		FROM engine.hosts;
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to list hosts from store: %w", err)
	}
	return h, nil
}

func (eng *Enginer) deleteHost(ctx context.Context, h host) error {
	_, err := h.db.ExecContext(ctx, `
		DELETE FROM engine.hosts
		WHERE host_id = $1;`, h.id)
	if err != nil {
		return fmt.Errorf("failed to delete host: %w", err)
	}
	return nil
}

// allocateHosts for a given mixer.
func (eng *Enginer) allocateHosts(ctx context.Context, m Mixer) error {
	allocations, err := eng.listHostsForMixer(ctx, m)
	if err != nil {
		return fmt.Errorf("failed to list engines: %w", err)
	}

	for i := 0; i < m.NumOfInstances-len(allocations); i++ {
		for _, host := range eng.hosts {
			// Find host without the ID.
			for _, allocation := range allocations {
				if *host != allocation {
					err = host.newAllocation(ctx, m)
					if err != nil {
						return fmt.Errorf("failed to allocate host: %w", err)
					}
					break
				}
			}
		}
	}
	return nil
}

func (h *host) newAllocation(ctx context.Context, m Mixer) error {
	h.db.ExecContext(ctx, `
		INSERT INTO engine.allocations (mixer_id, host_id)
		VALUES ($1, $2);
	`, m.ID, h.id)
	return nil
}

func (eng *Enginer) listHostsForMixer(ctx context.Context, m Mixer) ([]host, error) {
	hosts := []host{}
	err := eng.db.SelectContext(ctx, &hosts, `
		SELECT host_id
		FROM engine.allocations
		WHERE mixer_id = $1;
	`, m.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to list hosts: %w", err)
	}
	return hosts, nil
}
