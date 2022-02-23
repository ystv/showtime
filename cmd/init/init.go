package main

import (
	"context"
	"log"

	"github.com/ystv/showtime/db"
	"github.com/ystv/showtime/playout"
)

func main() {
	db, err := db.New()
	if err != nil {
		log.Fatalf("unable to create database: %+v", err)
	}

	_, err = db.ExecContext(context.Background(), playout.Schema)
	if err != nil {
		log.Fatalf("failed to create playout schema: %+v", err)
	}
	log.Println("successfully initialised showtime")
}
