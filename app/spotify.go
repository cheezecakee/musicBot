package app

import (
	"context"
	"discordBot/app/auth"
	"fmt"

	"github.com/zmb3/spotify/v2"
)

var Clients *auth.Clients

func SearchTrack(trackName string) (*spotify.FullTrack, error) {
	ctx := context.Background()
	results, err := Clients.Spotify.Search(ctx, trackName, spotify.SearchTypeTrack)
	if err != nil {
		return nil, fmt.Errorf("error searching Spotify: %v", err)
	}

	if results.Tracks.Total > 0 {
		return &results.Tracks.Tracks[0], nil
	}
	return nil, fmt.Errorf("no tracks found")
}
