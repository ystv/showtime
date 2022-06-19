package auth

import (
	"log"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/youtube/v3"
)

// NewYouTubeConfig creates a oauth2 config for YouTube.
func NewYouTubeConfig(b []byte) (*oauth2.Config, error) {
	// Temporary debug.
	log.Println("temporary debug start")
	log.Println(string(b))
	log.Println("temporary debug end")
	// If modifying these scopes, delete your previously saved credentials file.
	return google.ConfigFromJSON(b, youtube.YoutubeForceSslScope)
}
