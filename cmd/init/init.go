package main

import (
	"flag"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/pressly/goose/v3"

	"github.com/ystv/showtime/db"
)

func main() {
	// Load environment
	_ = godotenv.Load(".env")           // Load .env file for production
	_ = godotenv.Overload(".env.local") // Load .env.local for developing

	downOne := flag.Bool("down_one", false, "undo the last migration instead of upgrading - only use for development!")
	flag.Parse()

	dbConf := &db.Config{
		Host:                   os.Getenv("ST_DB_HOST"),
		Port:                   os.Getenv("ST_DB_PORT"),
		SSLMode:                os.Getenv("ST_DB_SSLMODE"),
		DBName:                 os.Getenv("ST_DB_DBNAME"),
		Username:               os.Getenv("ST_DB_USERNAME"),
		Password:               os.Getenv("ST_DB_PASSWORD"),
		SkipSchemaVersionCheck: true,
	}
	db, err := db.New(dbConf)
	if err != nil {
		log.Fatalf("unable to create database: %+v", err)
	}
	defer db.Close()

	if *downOne {
		if err := goose.Down(db.DB, "."); err != nil {
			log.Fatalf("unable to downgrade: %v", err)
		}
		return
	}

	if err := goose.Up(db.DB, "."); err != nil {
		log.Fatalf("unable to run migrations: %v", err)
	}

	log.Println("successfully initialised showtime")
}
