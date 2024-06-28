package app

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/zmb3/spotify/v2"
	"github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2/clientcredentials"
)

var spotifyClient *spotify.Client

func InitSpotify() {
	ctx := context.Background()
	config := &clientcredentials.Config{
		ClientID:     os.Getenv("SPOTIFY_ID"),
		ClientSecret: os.Getenv("SPOTIFY_SECRET"),
		TokenURL:     spotifyauth.TokenURL,
	}

	token, err := config.Token(ctx)
	if err != nil {
		log.Fatalf("couldn't get token: %v", err)
	}

	httpClient := spotifyauth.New().Client(ctx, token)
	spotifyClient = spotify.New(httpClient)

	// Log initialization success
	log.Println("Spotify client initialized successfully.")
}

func SearchTrack(ctx context.Context, trackName string) (*spotify.FullTrack, error) {
	results, err := spotifyClient.Search(ctx, trackName, spotify.SearchTypeTrack)
	if err != nil {
		return nil, err
	}

	if results.Tracks.Total > 0 {
		return &results.Tracks.Tracks[0], nil
	}
	return nil, fmt.Errorf("no tracks found")
}
