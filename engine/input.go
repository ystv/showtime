package engine

import (
	"context"
	"fmt"
)

type (
	// Input represents a source.
	Input struct {
		ID int `db:"input_id"`
	}
)

// GetInput retrieves an input.
func (eng *Enginer) GetInput(ctx context.Context, inputID int) (Input, error) {
	i := Input{}

	err := eng.db.GetContext(ctx, &i, `
		SELECT input_id
		FROM engine.inputs
		WHERE input_id = $1
	`, inputID)
	if err != nil {
		return Input{}, fmt.Errorf("failed to get input from store: %w", err)
	}

	return i, nil
}
