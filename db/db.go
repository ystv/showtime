package db

import (
	"fmt"
	"log"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/pressly/goose/v3"

	// Postgres driver
	_ "github.com/lib/pq"

	"github.com/ystv/showtime/db/migrations"
)

func init() {
	goose.SetBaseFS(migrations.Migrations)
	if err := goose.SetDialect("postgres"); err != nil {
		panic(fmt.Errorf("failed to set goose dialect: %w", err))
	}
}

// Config required to connect to database.
type Config struct {
	Host                   string
	Port                   string
	SSLMode                string
	DBName                 string
	Username               string
	Password               string
	SkipSchemaVersionCheck bool
	AutoInit               bool
}

// New creates a new database client.
func New(conf *Config) (*sqlx.DB, error) {
	// ref. https://pkg.go.dev/github.com/lib/pq#hdr-Connection_String_Parameters for formatting
	dbURI := fmt.Sprintf("host=%s port=%s sslmode=%s dbname=%s user=%s password='%s' application_name=ShowTime!",
		conf.Host, conf.Port, conf.SSLMode, conf.DBName, conf.Username, strings.ReplaceAll(conf.Password, "'", "\\'"))
	db, err := sqlx.Open("postgres", dbURI)
	if err != nil {
		return nil, fmt.Errorf("failed to open db: %w", err)
	}
	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("failed to ping db: %w", err)
	}

	if conf.AutoInit {
		log.Printf("auto-initialising database")
		if err := goose.Up(db.DB, "."); err != nil {
			return nil, fmt.Errorf("failed to run goose migrations: %w", err)
		}
	}

	if !conf.SkipSchemaVersionCheck {
		upToDate, err := migrations.IsUpToDate(db.DB)
		if err != nil {
			return nil, fmt.Errorf("failed to check if migrations are up to date: %w", err)
		}
		if !upToDate {
			return nil, fmt.Errorf("database not up to date, please run 'init' or set ST_DB_AUTO_INIT=true")
		}
	}

	return db, nil
}
