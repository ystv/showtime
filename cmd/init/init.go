package main

import (
	"context"
	"log"

	"github.com/ystv/showtime/auth"
	"github.com/ystv/showtime/db"
	"github.com/ystv/showtime/livestream"
	"github.com/ystv/showtime/mcr"
	"github.com/ystv/showtime/youtube"
)

func main() {
	db, err := db.New()
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
