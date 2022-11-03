package engine

import (
	"context"
	"fmt"

	"github.com/ystv/showtime/brave"
)

type (
	// Mixer represents a video compositor.
	Mixer struct {
		ID             int `db:"mixer_id"`
		Width          int `db:"width"`
		Height         int `db:"height"`
		NumOfInstances int `db:"num_of_instances"`

		eng     *Enginer
		hostIDs []int
	}

	// NewMixer parameters for a new mixer.
	NewMixer struct {
		Width          int
		Height         int
		NumOfInstances int
	}

	// Output is a stream out of the mixer.
	Output struct {
		ID      int    `db:"output_id"`
		Type    string `db:"type"`
		Address string `db:"address"`
	}
)

// NewMixer creates a new mixer
func (eng *Enginer) NewMixer(ctx context.Context, nm NewMixer) (Mixer, error) {
	m := Mixer{
		Width:          nm.Width,
		Height:         nm.Height,
		NumOfInstances: nm.NumOfInstances,
	}

	err := eng.db.GetContext(ctx, &m.ID, `
		INSERT INTO engine.mixers (width, height)
		VALUES ($1, $2)
		RETURNING mixer_id;
	`, nm.Width, nm.Height)
	if err != nil {
		return Mixer{}, fmt.Errorf("failed to insert mixer in store: %w", err)
	}

	err = eng.allocateHosts(ctx, m)
	if err != nil {
		return Mixer{}, fmt.Errorf("failed to allocate hosts: %w", err)
	}

	hostIDs := []int{}
	err = eng.db.SelectContext(ctx, &hostIDs, `
		SELECT host_id
		FROM engine.allocations
		WHERE mixer_id = $1;
	`, m.ID)
	if err != nil {
		return Mixer{}, fmt.Errorf("failed to get host allocations: %w", err)
	}

	for _, hostID := range hostIDs {
		err = eng.hosts[hostID].newMixer(ctx, m)
	}

	return m, nil
}

func (h *host) newMixer(ctx context.Context, m Mixer) error {
	bm, err := h.brave.NewMixer(ctx, brave.NewMixerParams{
		Width:  m.Width,
		Height: m.Height,
	})
	if err != nil {
		return fmt.Errorf("failed to create mixer in brave: %w", err)
	}

	_, err = h.db.ExecContext(ctx, `
		INSERT INTO engine.mappings (showtime_id, type, host_id, brave_id)
		VALUES ($1, $2, $3, $4);`, m.ID, "mixer", h.id, bm.ID)
	if err != nil {
		return fmt.Errorf("failed to insert mapping: %w", err)
	}
	return nil
}

// GetMixer retrieves a mixer instance.
func (eng *Enginer) GetMixer(ctx context.Context, mixerID int) (Mixer, error) {
	m := Mixer{}

	err := eng.db.GetContext(ctx, &m, `
		SELECT mixer_id, width, height, num_of_instances
		FROM engine.mixers
		WHERE mixer_id = $1;
	`, mixerID)
	if err != nil {
		return Mixer{}, fmt.Errorf("failed to get mixer from store: %w", err)
	}

	m.eng = eng

	err = m.eng.db.SelectContext(ctx, &m.hostIDs, `
		SELECT host_id
		FROM engine.engines
		WHERE mixer_id = $1;`, m.ID)
	if err != nil {
		return Mixer{}, fmt.Errorf("failed to get hosts: %w", err)
	}
	return m, nil
}

// CutToInput cuts a mixer on engines to an input.
func (m *Mixer) CutToInput(ctx context.Context, i Input) error {
	for _, hostID := range m.hostIDs {
		err := m.eng.hosts[hostID].cutMixerToInput(ctx, *m, i)
		if err != nil {
			return fmt.Errorf("failed to cut mixer to input: %w", err)
		}
	}
	return nil
}

func (h *host) cutMixerToInput(ctx context.Context, m Mixer, i Input) error {
	mixerID := 0
	err := h.db.GetContext(ctx, &mixerID, `
		SELECT brave_id
		FROM engine.mappings
		WHERE host_id = $1 AND showtime_id = $2;
	`, h.id, m.ID)
	if err != nil {
		return fmt.Errorf("failed to find mixer for host: %w", err)
	}

	inputID := 0
	err = h.db.GetContext(ctx, &inputID, `
		SELECT brave_id
		FROM engine.mappings
		WHERE host_id = $1 AND showtime_id = $2;
	`, h.id, i.ID)
	if err != nil {
		return fmt.Errorf("failed to find input for host: %w", err)
	}

	err = h.brave.CutMixerToInput(ctx, mixerID, inputID)
	if err != nil {
		return fmt.Errorf("failed to cut mixer to input in brave: %w", err)
	}

	_, err = h.db.ExecContext(ctx, `
		UPDATE engine.mixers
			SET program_input_id = $1
		WHERE host_id = $2 AND mixer_id = $3;
	`, inputID, h.id, mixerID)
	if err != nil {
		return fmt.Errorf("failed to update program input in store: %w", err)
	}
	return nil
}

// NewOutput creates an RTMP output of the mix.
func (m *Mixer) NewOutput(ctx context.Context, address string) (Output, error) {
	o := Output{
		Address: address,
		Type:    "rtmp",
	}

	err := m.eng.db.GetContext(ctx, &o.ID, `
		INSERT INTO engine.outputs (type, address, mixer_id)
		VALUES ($1, $2)
		RETURNING output_id;
	`, "rtmp", address)
	if err != nil {
		return Output{}, fmt.Errorf("failed to create output in store: %w", err)
	}

	for _, hostID := range m.hostIDs {
		braveMixerID := 0
		err = m.eng.db.GetContext(ctx, &braveMixerID, `
			SELECT brave_id
			FROM engine.mappings
			WHERE type = 'mixer' AND showtime_id = $1;
		`, m.ID)
		if err != nil {
			return Output{}, fmt.Errorf("failed to get mixer mapping: %w", err)
		}

		braveMixer, err := m.eng.hosts[hostID].brave.GetMixer(ctx, braveMixerID)
		if err != nil {
			return Output{}, fmt.Errorf("failed to get brave mixer: %w", err)
		}

		braveOutput, err := m.eng.hosts[hostID].brave.NewRTMPOutput(ctx, braveMixer, fmt.Sprintf("%s-%d", address, hostID))
		if err != nil {
			return Output{}, fmt.Errorf("failed to create new rtmp output: %w", err)
		}

		_, err = m.eng.db.ExecContext(ctx, `
			INSERT INTO engine.mappings (showtime_id, type, host_id, brave_id)
			VALUES ($1, $2, $3, $4);`, o.ID, "output", hostID, braveOutput.ID)
		if err != nil {
			return Output{}, fmt.Errorf("failed to insert mapping: %w", err)
		}
	}
	return o, nil
}
