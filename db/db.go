package db

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	// Postgres driver
	_ "github.com/lib/pq"
)

// Config required to connect to database.
type Config struct {
	Host     string
	Port     string
	SSLMode  string
	DBName   string
	Username string
	Password string
}

// New creates a new database client.
func New(conf *Config) (*sqlx.DB, error) {
	dbURI := fmt.Sprintf("host=%s port=%s sslmode=%s dbname=%s user=%s password=%s application_name=ShowTime!",
		conf.Host, conf.Port, conf.SSLMode, conf.DBName, conf.Username, conf.Password)
	db, err := sqlx.Open("postgres", dbURI)
	if err != nil {
		return nil, fmt.Errorf("failed to open db: %w", err)
	}
	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("failed to ping db: %w", err)
	}
	return db, nil
}
