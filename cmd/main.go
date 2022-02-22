package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/ystv/showtime/auth"
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
	config, err := google.ConfigFromJSON(b, youtube.YoutubeUploadScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %+v", err)
	}

	auth := auth.NewAuther(config)

	server := &http.Server{
		Addr:         ":8080",
		Handler:      handlers.New(auth).GetHandlers(),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}
	log.Printf("Starting HTTP Server. Listening at %q", server.Addr)
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		log.Printf("%v", err)
	} else {
		log.Println("Server closed!")
	}
}
