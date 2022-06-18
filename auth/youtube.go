package auth

import (
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/youtube/v3"
)

func NewYouTubeConfig(b []byte) (*oauth2.Config, error) {
	// If modifying these scopes, delete your previously saved token.json.
	return google.ConfigFromJSON(b, youtube.YoutubeForceSslScope)
}
