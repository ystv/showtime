package main

import (
	"context"
	"embed"
	"io/fs"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"

	"github.com/ystv/showtime/auth"
	"github.com/ystv/showtime/brave"
	"github.com/ystv/showtime/db"
	"github.com/ystv/showtime/handlers"
	"github.com/ystv/showtime/livestream"
	"github.com/ystv/showtime/mcr"
	"github.com/ystv/showtime/youtube"
)

//go:embed public/*
var content embed.FS

// Config for ShowTime!
type Config struct {
	livestream livestream.Config
	mcr        *mcr.Config
	brave      brave.Config
	handlers   *handlers.Config
	auth       *auth.Config
	db         *db.Config
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

	autoInit, _ := strconv.ParseBool(os.Getenv("ST_DB_AUTO_INIT"))

	conf := Config{
		livestream: livestream.Config{
			IngestAddress: os.Getenv("ST_INGEST_ADDR"),
		},
		mcr: &mcr.Config{
			BaseServeURL:  os.Getenv("ST_BASE_SERVE_ADDR"),
			OutputAddress: os.Getenv("ST_OUTPUT_ADDR"),
		},
		brave: brave.Config{
			Endpoint: os.Getenv("ST_BRAVE_ADDR"),
		},
		handlers: &handlers.Config{
			Debug:           debug,
			StateCookieName: "state-token",
			DomainName:      os.Getenv("ST_DOMAIN_NAME"),
			IngestAddress:   os.Getenv("ST_INGEST_ADDR"),
			JWTSigningKey:   os.Getenv("ST_SIGNING_KEY"),
		},
		auth: &auth.Config{
			CredentialsPath: os.Getenv("ST_CRED_PATH"),
		},
		db: &db.Config{
			Host:     os.Getenv("ST_DB_HOST"),
			Port:     os.Getenv("ST_DB_PORT"),
			SSLMode:  os.Getenv("ST_DB_SSLMODE"),
			DBName:   os.Getenv("ST_DB_DBNAME"),
			Username: os.Getenv("ST_DB_USERNAME"),
			Password: os.Getenv("ST_DB_PASSWORD"),
			AutoInit: autoInit,
		},
	}

	db, err := db.New(conf.db)
	if err != nil {
		log.Fatalf("unable to create database: %+v", err)
	}

	if conf.auth.CredentialsPath == "" {
		conf.auth.CredentialsPath = "credentials"
	}
	b, err := os.ReadFile(conf.auth.CredentialsPath + "/youtube.json")
	if err != nil {
		log.Fatalf("unable to read client secret file: %+v", err)
	}

	ytConfig, err := auth.NewYouTubeConfig(b)
	if err != nil {
		log.Fatalf("failed to create youtube config: %+v", err)
	}
	auth := auth.NewAuther(db, ytConfig)

	brave, err := brave.New(conf.brave)
	if err != nil {
		log.Fatalf("failed to create brave client: %+v", err)
	}
	mcr, err := mcr.NewMCR(conf.mcr, db, brave)
	if err != nil {
		log.Fatalf("failed to create mcr: %+v", err)
	}
	yt, err := youtube.New(context.Background(), db, auth)
	if err != nil {
		log.Fatalf("failed to create youtube client: %+v", err)
	}
	ls := livestream.New(conf.livestream, db, mcr, yt)

	templatesFS, err := fs.Sub(content, "public/templates")
	if err != nil {
		log.Fatalf("template files failed: %+v", err)
	}
	templates, err := handlers.NewTemplater(templatesFS)
	if err != nil {
		log.Fatalf("failed to create templater: %w", err)
	}

	h := handlers.New(conf.handlers, auth, ls, mcr, yt, templates)

	h.Start()
}
