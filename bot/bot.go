package bot

import (
	"context"
	"discordBot/app"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/zmb3/spotify/v2"
	"google.golang.org/api/youtube/v3"
)

var BotToken string

func checkNilErr(e error) {
	if e != nil {
		log.Fatal(e)
	}
}

func Run() {
	// Create a session
	discord, err := discordgo.New("Bot " + BotToken)
	checkNilErr(err)

	// Add an event handler
	discord.AddHandler(newMessage)

	// Open session
	err = discord.Open()
	if err != nil {
		log.Fatal(err)
	}
	defer discord.Close()

	fmt.Println("Bot running...")
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
}

func newMessage(discord *discordgo.Session, message *discordgo.MessageCreate) {
	// Prevent bot from responding to its own messages
	if message.Author.ID == discord.State.User.ID {
		return
	}

	// Parse and handle commands
	if err := handleCommand(discord, message); err != nil {
		log.Println("Command error:", err)
	}
}

func handleCommand(discord *discordgo.Session, message *discordgo.MessageCreate) error {
	// Get the channel information
	channel, err := discord.State.Channel(message.ChannelID)
	if err != nil {
		return fmt.Errorf("error getting channel: %v", err)
	}

	// Get the guild information
	guild, err := discord.State.Guild(channel.GuildID)
	if err != nil {
		return fmt.Errorf("error getting guild: %v", err)
	}

	switch {
	case strings.Contains(message.Content, "!hello"):
		return handleHelloCommand(discord, message)
	case strings.Contains(message.Content, "!bye"):
		return handleByeCommand(discord, message)
	case strings.Contains(message.Content, "!join"):
		_, err = handleJoinCommand(discord, message, guild)
		if err != nil {
			return err
		}
	case strings.Contains(message.Content, "!leave"):
		handleLeaveCommand(discord, message, guild)
	case strings.Contains(message.Content, "!play"):
		track, err := searchSongSpotify(discord, message)
		if err != nil {
			return err
		}
		if err = searchSongYoutube(discord, message, track); err != nil {
			return err
		}
	default:
		// Handle unknown commands or ignore non-command messages
		return nil
	}
	return nil
}

func handleHelloCommand(discord *discordgo.Session, message *discordgo.MessageCreate) error {
	discord.ChannelMessageSend(message.ChannelID, "Hello World😃")
	return nil
}

func handleByeCommand(discord *discordgo.Session, message *discordgo.MessageCreate) error {
	discord.ChannelMessageSend(message.ChannelID, "Good bye👋")
	return nil
}

func handleJoinCommand(discord *discordgo.Session, message *discordgo.MessageCreate, guild *discordgo.Guild) (*discordgo.VoiceConnection, error) {
	// Find the voice state of the user
	voiceState, err := getVoiceState(guild, message.Author.ID)
	if err != nil {
		discord.ChannelMessageSend(message.ChannelID, "You must be in a voice channel to use this command.")
		return nil, fmt.Errorf("user not in a voice channel")
	}

	// Join the voice channel
	voice, err := discord.ChannelVoiceJoin(guild.ID, voiceState.ChannelID, false, false)
	if err != nil {
		discord.ChannelMessageSend(message.ChannelID, "Failed to join the voice channel.")
		return nil, fmt.Errorf("failed to join voice channel: %v", err)
	}

	discord.ChannelMessageSend(message.ChannelID, "Joined the voice channel!")
	return voice, nil
}

func handleLeaveCommand(discord *discordgo.Session, message *discordgo.MessageCreate, guild *discordgo.Guild) {
	voiceConnection, _ := handleJoinCommand(discord, message, guild)
	voiceConnection.Disconnect()
	discord.ChannelMessageSend(message.ChannelID, "Left the voice channel!")
}

func getVoiceState(guild *discordgo.Guild, userID string) (*discordgo.VoiceState, error) {
	for _, vs := range guild.VoiceStates {
		if vs.UserID == userID {
			return vs, nil
		}
	}
	return nil, fmt.Errorf("user not in a voice channel")
}

// func handlePlay(discord *discordgo.Session, message *discordgo.MessageCreate) error {
// 	return nil
// }

func playVideo(video *youtube.SearchResult) error {
	fmt.Println("Playing video:", video.Snippet.Title)
	return nil
}

func searchSongSpotify(discord *discordgo.Session, message *discordgo.MessageCreate) (*spotify.FullTrack, error) {
	ctx := context.Background()
	trackName := strings.TrimPrefix(message.Content, "!play")
	track, err := app.SearchTrack(ctx, trackName)
	if err != nil {
		discord.ChannelMessageSend(message.ChannelID, "Could not find the song")
		return nil, err
	}
	discord.ChannelMessageSend(message.ChannelID, fmt.Sprintf("Found sone %s by %s on Spotify", trackName, track.Artists[0].Name))
	return track, nil
}

func searchSongYoutube(discord *discordgo.Session, message *discordgo.MessageCreate, track *spotify.FullTrack) error {
	ctx := context.Background()
	video, err := app.SearchVideo(ctx, fmt.Sprintf("%s by %s", track.Name, track.Artists[0].Name))
	if err != nil {
		discord.ChannelMessageSend(message.ChannelID, "Could not find the song on YouTube")
		return err
	}
	discord.ChannelMessageSend(message.ChannelID, fmt.Sprintf("Found on YouTube: %s", video.Snippet.Title))
	return nil
}
