package main

import (
	"embed"
	"io/fs"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/ystv/showtime/auth"
	"github.com/ystv/showtime/db"
	"github.com/ystv/showtime/handlers"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/youtube/v3"
)

//go:embed public/*
var content embed.FS

type Config struct {
	handlers *handlers.Config
}

func main() {
	// Load environment
	godotenv.Load(".env")           // Load .env file for production
	godotenv.Overload(".env.local") // Load .env.local for developing

	b, err := os.ReadFile("credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %+v", err)
	}

	conf := Config{
		handlers: &handlers.Config{
			DomainName:    os.Getenv("ST_DOMAIN_NAME"),
			IngestAddress: os.Getenv("ST_INGEST_ADDR"),
		},
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, youtube.YoutubeForceSslScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %+v", err)
	}

	db, err := db.New()
	if err != nil {
		log.Fatalf("unable to create database: %+v", err)
	}

	auth := auth.NewAuther(config)

	templatesFS, err := fs.Sub(content, "public/templates")
	if err != nil {
		log.Fatalf("template files failed: %+v", err)
	}
	templates, err := handlers.NewTemplater(templatesFS)
	if err != nil {
		log.Fatalf("failed to create templater: %w", err)
	}

	h := handlers.New(conf.handlers, db, auth, templates)

	h.Start()
}
