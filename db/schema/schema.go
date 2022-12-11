// Package schema handles upgrading the database schema to the latest version.
// It's a separate package from `db` to prevent an import cycle.
package schema

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/jmoiron/sqlx"

	"github.com/ystv/showtime/auth"
	"github.com/ystv/showtime/db"
	"github.com/ystv/showtime/livestream"
	"github.com/ystv/showtime/mcr"
	"github.com/ystv/showtime/youtube"
)

var schemata = map[string]db.VersionedSchema{
	"core":       db.Schema,
	"livestream": livestream.Schema,
	"mcr":        mcr.Schema,
	"auth":       auth.Schema,
	"youtube":    youtube.Schema,
}

var HighestSchemaVersion uint16

func init() {
	for _, schema := range schemata {
		for version := range schema {
			if version > HighestSchemaVersion {
				HighestSchemaVersion = version
			}
		}
	}
}

func upgradeDatabaseTo(db *sqlx.DB, v uint16) error {
	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	for name, schema := range schemata {
		for sv, statements := range schema {
			if sv != v {
				continue
			}
			if _, err := tx.Exec(statements); err != nil {
				return fmt.Errorf("failed to execute schema statements for %q version %d: %w", name, v, err)
			}
		}
	}
	_, err = tx.Exec(`INSERT INTO schema_versions (version) VALUES ($1)`, v)
	if err != nil {
		return fmt.Errorf("failed to save new schema version: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit upgrade transaction: %w", err)
	}
	return nil
}

// UpgradeDatabase applies any migrations necessary to bring the database to the latest version.
func UpgradeDatabase(db *sqlx.DB) error {
	currentVersion, err := GetCurrentSchemaVersion(db)
	if err != nil {
		return err
	}

	if currentVersion > HighestSchemaVersion {
		return fmt.Errorf("database schema version %d is newer than the highest supported version %d", currentVersion, HighestSchemaVersion)
	}
	if currentVersion == HighestSchemaVersion {
		return nil // nothing to do
	}

	for i := currentVersion + 1; i <= HighestSchemaVersion; i++ {
		log.Printf("upgrading database to schema version %d", i)
		if err := upgradeDatabaseTo(db, i); err != nil {
			return fmt.Errorf("failed to upgrade database to version %d: %w", i, err)
		}
	}
	log.Printf("upgraded database to version %d", HighestSchemaVersion)
	return nil
}

// GetCurrentSchemaVersion returns the current version of the database schema. For details on the versioning scheme,
// see the documentation for db.VersionedSchema.
func GetCurrentSchemaVersion(db *sqlx.DB) (uint16, error) {
	row := db.QueryRowx(`SELECT MAX(version) AS v FROM schema_versions`)
	if err := row.Err(); err != nil {
		if strings.HasSuffix(err.Error(), `pq: relation "schema_versions" does not exist`) {
			return 0, nil
		}
		return 0, fmt.Errorf("failed to query current schema version: %w", err)
	}
	var currentVersion uint16
	if err := row.Scan(&currentVersion); err != nil {
		return 0, fmt.Errorf("failed to determine current schema version: %w", err)
	}
	return currentVersion, nil
}
