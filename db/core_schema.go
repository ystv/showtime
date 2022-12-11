package db

// Schema is the database schema for the ShowTime core.
var Schema = VersionedSchema{
	1: `
	CREATE TABLE schema_versions (
	    version SMALLINT NOT NULL PRIMARY KEY,
	    upgraded_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);
	`,
}
