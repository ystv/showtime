package migrations

import (
	"database/sql"
	"embed"
	"fmt"
	"math"

	"github.com/pressly/goose/v3"
)

//go:embed *.sql
var Migrations embed.FS

// IsUpToDate checks if the given database has all the migrations applied. Returns (false, nil) if not, or
// (false, err) if something went wrong in the process.
// NB: this function assumes goose.SetBaseFS has been called.
func IsUpToDate(db *sql.DB) (bool, error) {
	// migration version
	migrations, err := goose.CollectMigrations(".", 0, math.MaxInt)
	if err != nil {
		return false, fmt.Errorf("failed to collect migrations: %w", err)
	}
	last, err := migrations.Last()
	if err != nil {
		return false, fmt.Errorf("failed to get last migration: %w", err)
	}

	// db version
	version, err := goose.GetDBVersion(db)
	if err != nil {
		return false, fmt.Errorf("failed to get db version: %w", err)
	}

	return version == last.Version, nil
}
