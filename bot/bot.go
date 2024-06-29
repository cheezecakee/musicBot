package bot

import (
	"context"
	"discordBot/app"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/zmb3/spotify/v2"
	_ "google.golang.org/api/youtube/v3"
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
		err := handlePlay(discord, message)
		if err != nil {
			return err
		}
	default:
		// Handle unknown commands or ignore non-command messages
		return nil
	}
	return nil
}

func handleHelloCommand(discord *discordgo.Session, message *discordgo.MessageCreate) error {
	discord.ChannelMessageSend(message.ChannelID, "Hello Kosta😃")
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

func handlePlay(discord *discordgo.Session, message *discordgo.MessageCreate) error {
	// Search for the song on Spotify
	track, err := searchSongSpotify(discord, message)
	if err != nil {
		return nil
	}

	// Search for the video on YouTube
	videoID, err := searchSongYoutube(discord, message, track)
	if err != nil {
		return err
	}

	// Join the voice channel
	vc, err := handleJoinCommand(discord, message, discord.State.Guilds[0])
	if err != nil {
		return err
	}

	// Dowandload and convert the video to audio
	err = downloadConvertVideo(videoID, "/tmp/audio.mp3")
	if err != nil {
		discord.ChannelMessageSend(message.ChannelID, "Failed to download or convert the video")
	}

	return playAudio(vc, "/tmp/audio.mp3")
}

func playAudio(vc *discordgo.VoiceConnection, audioFilePath string) error {
	cmd := exec.Command("ffmpeg", "-i", audioFilePath, "-f", "s16le", "ar", "48000", "-ac", "2", "pipe:1")
	ffmpegOut, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create ffmpeg stdout pipe: %v", err)
	}

	err = cmd.Start()
	if err != nil {
		return fmt.Errorf("failed to start ffmpeg: %v", err)
	}

	vc.Speaking(true)
	for {
		buf := make([]byte, 960)
		n, err := ffmpegOut.Read(buf)
		if err != nil {
			break
		}
		vc.OpusSend <- buf[:n]
	}
	vc.Speaking(false)

	err = cmd.Wait()
	if err != nil {
		return fmt.Errorf("ffmpeg command failed: %v", err)
	}
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

func searchSongYoutube(discord *discordgo.Session, message *discordgo.MessageCreate, track *spotify.FullTrack) (string, error) {
	ctx := context.Background()
	video, err := app.SearchVideo(ctx, fmt.Sprintf("%s by %s", track.Name, track.Artists[0].Name))
	if err != nil {
		discord.ChannelMessageSend(message.ChannelID, "Could not find the song on YouTube")
		return "", err
	}
	discord.ChannelMessageSend(message.ChannelID, fmt.Sprintf("Found on YouTube: %s", video))
	return video, nil
}

func downloadConvertVideo(videoID string, outputPath string) error {
	// Construct the download URL
	url := fmt.Sprintf("https://www.youtube.com/watch?v=%s", videoID)

	// Download the video
	err := downloadFile(url, "/tmp/video.webm")
	if err != nil {
		return err
	}

	// Convert the video to MP3"
	cmd := exec.Command("ffmpeg", "-i", "/tmp/video.webm", "-vn", "-ar", "44100", "-ac", "2", "-b:a", "192k", outputPath)
	err = cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

func downloadFile(url, outputPath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}
