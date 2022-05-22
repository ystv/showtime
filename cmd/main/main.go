package main

import (
	"embed"
	"io/fs"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/ystv/showtime/auth"
	"github.com/ystv/showtime/brave"
	"github.com/ystv/showtime/channel"
	"github.com/ystv/showtime/db"
	"github.com/ystv/showtime/handlers"
	"github.com/ystv/showtime/playout"
	"github.com/ystv/showtime/youtube"
)

//go:embed public/*
var content embed.FS

// Config for ShowTime!
type Config struct {
	playout  playout.Config
	brave    brave.Config
	handlers *handlers.Config
}

func main() {
	// Load environment
	godotenv.Load(".env")           // Load .env file for production
	godotenv.Overload(".env.local") // Load .env.local for developing

	// Check if debugging
	debug, err := strconv.ParseBool(os.Getenv("ST_DEBUG"))
	if err != nil {
		debug = false
		os.Setenv("ST_DEBUG", "false")
	}
	if debug {
		log.Println("Debug Mode - Disabled auth - pls don't run in production")
	}

	conf := Config{
		playout: playout.Config{
			IngestAddress: os.Getenv("ST_INGEST_ADDR"),
		},
		brave: brave.Config{
			Endpoint: os.Getenv("ST_BRAVE_ADDR"),
		},
		handlers: &handlers.Config{
			Debug:           debug,
			StateCookieName: "state-token",
			DomainName:      os.Getenv("ST_DOMAIN_NAME"),
			JWTSigningKey:   os.Getenv("ST_SIGNING_KEY"),
		},
	}

	b, err := os.ReadFile("credentials.json")
	if err != nil {
		log.Fatalf("unable to read client secret file: %+v", err)
	}

	ytConfig, err := auth.NewYouTubeConfig(b)
	if err != nil {
		log.Fatalf("failed to create youtube config: %+v", err)
	}
	auth := auth.NewAuther(ytConfig)

	db, err := db.New()
	if err != nil {
		log.Fatalf("unable to create database: %+v", err)
	}

	brave, err := brave.New(conf.brave)
	if err != nil {
		log.Fatalf("failed to create brave client: %+v", err)
	}
	mcr := channel.NewMCR(db, brave)
	yt, err := youtube.New(db, auth)
	if err != nil {
		log.Fatalf("failed to create youtube client: %+v", err)
	}
	play := playout.New(conf.playout, db, yt)

	templatesFS, err := fs.Sub(content, "public/templates")
	if err != nil {
		log.Fatalf("template files failed: %+v", err)
	}
	templates, err := handlers.NewTemplater(templatesFS)
	if err != nil {
		log.Fatalf("failed to create templater: %w", err)
	}

	h := handlers.New(conf.handlers, auth, play, mcr, yt, templates)

	h.Start()
}
