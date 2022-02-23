package db

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

func New() (*sqlx.DB, error) {
	db, err := sqlx.Open("sqlite3", "./showtime.db")
	if err != nil {
		return nil, fmt.Errorf("failed to open db: %w", err)
	}
	return db, nil
}
