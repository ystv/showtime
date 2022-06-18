package auth

import (
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/youtube/v3"
)

// NewYouTubeConfig creates a oauth2 config for YouTube.
func NewYouTubeConfig(b []byte) (*oauth2.Config, error) {
	// If modifying these scopes, delete your previously saved credentials file.
	return google.ConfigFromJSON(b, youtube.YoutubeForceSslScope)
}
