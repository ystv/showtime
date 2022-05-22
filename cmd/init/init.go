package main

import (
	"context"
	"log"

	"github.com/ystv/showtime/channel"
	"github.com/ystv/showtime/db"
	"github.com/ystv/showtime/playout"
	"github.com/ystv/showtime/youtube"
)

func main() {
	db, err := db.New()
	if err != nil {
		log.Fatalf("unable to create database: %+v", err)
	}

	ctx := context.Background()

	_, err = db.ExecContext(ctx, playout.Schema)
	if err != nil {
		log.Fatalf("failed to create playout schema: %+v", err)
	}
	_, err = db.ExecContext(ctx, channel.Schema)
	if err != nil {
		log.Fatalf("failed to create channel schema: %+v", err)
	}
	_, err = db.ExecContext(ctx, youtube.Schema)
	if err != nil {
		log.Fatalf("failed to create youtube schema: %+v", err)
	}

	log.Println("successfully initialised showtime")
}
