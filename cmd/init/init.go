package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"

	"github.com/ystv/showtime/db"
	"github.com/ystv/showtime/db/schema"
)

func main() {
	// Load environment
	_ = godotenv.Load(".env")           // Load .env file for production
	_ = godotenv.Overload(".env.local") // Load .env.local for developing

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
	defer db.Close()

	if err := schema.UpgradeDatabase(db); err != nil {
		log.Fatal(err)
	}

	log.Println("successfully initialised showtime")
}
