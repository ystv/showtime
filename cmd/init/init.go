package main

import (
	"context"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/ystv/showtime/auth"
	"github.com/ystv/showtime/db"
	"github.com/ystv/showtime/livestream"
	"github.com/ystv/showtime/mcr"
	"github.com/ystv/showtime/youtube"
)

func main() {
	// Load environment
	godotenv.Load(".env")           // Load .env file for production
	godotenv.Overload(".env.local") // Load .env.local for developing

	dbConf := &db.Config{
		Host:     os.Getenv("ST_DB_HOST"),
		Port:     os.Getenv("ST_DB_PORT"),
		SSLMode:  os.Getenv("ST_DB_SSLMODE"),
		DBName:   os.Getenv("ST_DB_DBNAME"),
		Username: os.Getenv("ST_DB_USERNAME"),
		Password: os.Getenv("ST_DB_PASSWORD"),
	}
	db, err := db.New(dbConf)
	if err != nil {
		log.Fatalf("unable to create database: %+v", err)
	}

	ctx := context.Background()

	_, err = db.ExecContext(ctx, livestream.Schema)
	if err != nil {
		log.Fatalf("failed to create livestream schema: %+v", err)
	}
	_, err = db.ExecContext(ctx, mcr.Schema)
	if err != nil {
		log.Fatalf("failed to create channel schema: %+v", err)
	}
	_, err = db.ExecContext(ctx, auth.Schema)
	if err != nil {
		log.Fatalf("failed to create auth schema: %+v", err)
	}
	_, err = db.ExecContext(ctx, youtube.Schema)
	if err != nil {
		log.Fatalf("failed to create youtube schema: %+v", err)
	}

	log.Println("successfully initialised showtime")
}
