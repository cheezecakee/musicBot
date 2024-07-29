package auth

import (
	"context"
	"os"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"github.com/zmb3/spotify/v2"
	"github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2/clientcredentials"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

type Token struct {
	DISCORD_BOT    string
	SPOTIFY_ID     string
	SPOTIFY_SECRET string
	YOUTUBE_API    string
}

type Clients struct {
	Discord *discordgo.Session
	Spotify *spotify.Client
	Youtube *youtube.Service
}

func LoadTokens() (*Token, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, err
	}

	tokens := &Token{
		DISCORD_BOT:    os.Getenv("DISCORD_BOT"),
		SPOTIFY_ID:     os.Getenv("SPOTIFY_ID"),
		SPOTIFY_SECRET: os.Getenv("SPOTIFY_SECRET"),
		YOUTUBE_API:    os.Getenv("YOUTUBE_API"),
	}

	return tokens, nil
}

func InitClients(tokens *Token) (*Clients, error) {
	discord, err := discordgo.New("Bot " + tokens.DISCORD_BOT)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()

	spotifyConfig := &clientcredentials.Config{
		ClientID:     tokens.SPOTIFY_ID,
		ClientSecret: tokens.SPOTIFY_SECRET,
		TokenURL:     spotifyauth.TokenURL,
	}

	spotifyToken, err := spotifyConfig.Token(ctx)
	if err != nil {
		return nil, err
	}

	spotifyClient := spotify.New(spotifyauth.New().Client(ctx, spotifyToken))

	youtubeService, err := youtube.NewService(ctx, option.WithAPIKey(tokens.YOUTUBE_API))
	if err != nil {
		return nil, err
	}

	clients := &Clients{
		Discord: discord,
		Spotify: spotifyClient,
		Youtube: youtubeService,
	}

	return clients, nil
}
