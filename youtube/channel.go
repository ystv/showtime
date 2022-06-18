package youtube

import (
	"context"
	"fmt"
)

// ChannelInfo is a brief summary of a channel.
type ChannelInfo struct {
	AccountID   int
	Name        string
	Description string
	Link        string
	Image       string
}

// About returns a brief summary of all connected account's channels.
func (y *YouTube) About(ctx context.Context) ([]ChannelInfo, error) {
	info := []ChannelInfo{}
	for _, yt := range y.youtubers {
		chInfo, err := yt.About(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to about info on channel: %w", err)
		}
		info = append(info, chInfo...)
	}
	return info, nil
}

// About returns a brief summary of channels.
func (y *YouTuber) About(ctx context.Context) ([]ChannelInfo, error) {
	info := []ChannelInfo{}
	chs, err := y.yt.Channels.List([]string{"snippet"}).Mine(true).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to list channels: %w", err)
	}
	for _, ch := range chs.Items {
		info = append(info, ChannelInfo{
			AccountID:   y.accountID,
			Name:        ch.Snippet.Title,
			Description: ch.Snippet.Description,
			Link:        "https://youtube.com/channel/" + ch.Id,
			Image:       ch.Snippet.Thumbnails.Default.Url,
		})
	}
	return info, nil
}
