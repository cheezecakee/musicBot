package bot

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/bwmarrin/discordgo"
)

type URLHandlerFunc func(bot *Bot, urlID string, i *discordgo.InteractionCreate)

var urlHandlers = map[string]URLHandlerFunc{
	"https://open.spotify.com/playlist/": handleSpotifyPlaylist,
	"https://open.spotify.com/track/":    handleSpotifyTrack,
	"https://www.youtube.com/watch":      handleYouTubeVideo,
}

// Register handlers for each URL type
func handleSpotifyPlaylist(bot *Bot, urlID string, i *discordgo.InteractionCreate) {
	playlistID := extractID(urlID)
	if playlistID == "" {
		bot.sendFollowUp(i, "Invalid Spotify playlist URL")
		return
	}
	bot.Player.FindPlaylistSpotify(playlistID)
	bot.sendFollowUp(i, "Playlist added to queue")
}

func handleSpotifyTrack(bot *Bot, urlID string, i *discordgo.InteractionCreate) {
	bot.Player.Name = urlID
	bot.Player.Find()
	bot.sendFollowUp(i, fmt.Sprintf("Added to queue: %s by %s", bot.Player.Track.Name, bot.Player.Track.Artists[0].Name))
}

func handleYouTubeVideo(bot *Bot, urlID string, i *discordgo.InteractionCreate) {
	videoID := extractID(urlID)

	track := bot.Player.FindYoutube(videoID)

	bot.sendFollowUp(i, fmt.Sprintf("Added to queue: %s ", track))
}

// Helper function to extract url from user play command input
func extractID(urlID string) string {
	parsedURL, err := url.Parse(urlID)
	if err != nil {
		return ""
	}
	if strings.Contains(urlID, "spotify.com/playlist/") {
		return strings.TrimPrefix(parsedURL.Path, "/playlist/")
	}
	if strings.Contains(urlID, "youtube.com/watch") {
		return parsedURL.Query().Get("v")
	}
	return ""
}

func isSpotifyPlaylistURL(urlID string) bool {
	return strings.HasPrefix(urlID, "https://open.spotify.com/playlist/")
}

func isYouTubeVideoURL(urlID string) bool {
	return strings.HasPrefix(urlID, "https://www.youtube.com/watch")
}
