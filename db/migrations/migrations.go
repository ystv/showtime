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
	migrations, err := goose.CollectMigrations(".", 0, math.MaxInt64)
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

	if version > last.Version {
		return false, fmt.Errorf("db version is greater than the latest migration version, this ShowTime binary is outdated")
	}

	return version == last.Version, nil
}
