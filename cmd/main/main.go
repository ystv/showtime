package main

import (
	"log"
	"os"

	"github.com/ystv/showtime/auth"
	"github.com/ystv/showtime/db"
	"github.com/ystv/showtime/handlers"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/youtube/v3"
)

func main() {
	b, err := os.ReadFile("credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %+v", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, youtube.YoutubeForceSslScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %+v", err)
	}

	db, err := db.New()
	if err != nil {
		log.Fatalf("unable to create database: %w", err)
	}

	auth := auth.NewAuther(config)

	h := handlers.New(db, auth)

	h.Start()
}
