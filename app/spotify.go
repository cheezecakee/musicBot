package app

import (
	"context"
	"discordBot/app/auth"
	"fmt"
	"net/url"
	"strings"

	"github.com/zmb3/spotify/v2"
)

var Clients *auth.Clients

func SearchTrack(query string) (*spotify.FullTrack, error) {
	ctx := context.Background()

	// Check if the query is a Spotify URL
	if strings.HasPrefix(query, "https://open.spotify.com/track/") {
		parsedURL, err := url.Parse(query)
		if err != nil {
			return nil, fmt.Errorf("invalid URL: %v", err)
		}
		segments := strings.Split(parsedURL.Path, "/")
		if len(segments) > 2 {
			trackID := segments[2]
			track, err := Clients.Spotify.GetTrack(ctx, spotify.ID(trackID))
			if err != nil {
				return nil, fmt.Errorf("error fetching track by ID: %v", err)
			}
			return track, nil
		}
	}

	// Search for tracks by name
	results, err := Clients.Spotify.Search(ctx, query, spotify.SearchTypeTrack)
	if err != nil {
		return nil, fmt.Errorf("error searching Spotify: %v", err)
	}

	// Check if any tracks are found by name
	if results.Tracks.Total > 0 {
		return &results.Tracks.Tracks[0], nil
	}

	// If no tracks are found by name, return an error
	return nil, fmt.Errorf("no tracks found")
}
