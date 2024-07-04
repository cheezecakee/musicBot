package bot

import (
	"discordBot/app"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/jonas747/dca"
)

func (bot *Bot) streamAudio() error {
	err := bot.handleJoinCommand()
	if err != nil {
		return (err)
	}
	bot.VoiceConnection.Speaking(true)

	bot.sendMessage(fmt.Sprintf("Now playing: %s by %s", bot.pl.track.Name, bot.pl.track.Artists[0].Name))

	bot.convertToDCA()

	file, err := os.Open(bot.pl.path)
	if err != nil {
		fmt.Println("Error opening dca file :", err)
		return err
	}

	// Create a DCA decoder to read from the encoded DCA file
	decoder := dca.NewDecoder(file)

	// Read and stream each Opus frame to Discord
	for {
		frame, err := decoder.OpusFrame()
		if err != nil {
			if err != io.EOF {
				return nil
			}
			break // End of audio stream
		}

		// Send the Opus frame to Discord
		select {
		case bot.VoiceConnection.OpusSend <- frame:
		case <-time.After(time.Second):
			return nil
		}
	}

	bot.VoiceConnection.Disconnect()
	err = bot.cleanUpDir()
	if err != nil {
		fmt.Printf("Failed to clean up temp directory: %v\n", err)
	}

	fmt.Println("Audio stream completed successfully")
	return nil
}

func (bot *Bot) convertToDCA() error {
	// Construct the ffmpeg command to convert the audio file to DCA format
	cmd := exec.Command("ffmpeg", "-i", bot.pl.output, "-f", "s16le", "-ac", "2", "-ar", "48000", "-acodec", "pcm_s16le", "-")

	// Create path with song name
	bot.pl.path = "./player/" + bot.pl.track.Name + ".dca"

	// Pipe the output to dca
	dcaCmd := exec.Command("dca")
	dcaCmd.Stdin, _ = cmd.StdoutPipe()
	dcaCmd.Stdout, _ = os.Create(bot.pl.path)

	// Start the ffmpeg command
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start ffmpeg command: %v", err)
	}

	// Start the dca command
	if err := dcaCmd.Start(); err != nil {
		return fmt.Errorf("failed to start dca command: %v", err)
	}

	// Wait for ffmpeg to finish
	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("ffmpeg command failed: %v", err)
	}

	// Wait for dca to finish
	if err := dcaCmd.Wait(); err != nil {
		return fmt.Errorf("dca command failed: %v", err)
	}

	fmt.Println(bot.pl.output)

	fmt.Println("Conversion to DCA completed successfully")
	return nil
}

func (bot *Bot) downloadAudio() error {
	// Search for the track on Spotify
	track, err := app.SearchTrack(bot.ctx, bot.pl.trackName)
	if err != nil {
		return fmt.Errorf("error searching track: %v", err)
	}
	bot.pl.track = track
	// Use the track name and artist for the Youtube search query
	query := fmt.Sprintf("%s %s official", track.Name, track.Artists[0].Name)
	bot.pl.videoID, err = app.SearchVideo(bot.ctx, query)
	if err != nil {
		return fmt.Errorf("error search video: %v", err)
	}

	// Convert the video to audio
	err = bot.convertVideo()
	if err != nil {
		return fmt.Errorf("error converting video: %v", err)
	}

	// fmt.Println("video ID:", bot.pl.videoID)
	return nil
}

func (bot *Bot) convertVideo() error {
	// Construst the download URL
	url := fmt.Sprintf("https://www.youtube.com/watch?v=%s", bot.pl.videoID)

	youtubeDownloader, err := exec.LookPath("yt-dlp")
	if err != nil {
		fmt.Println("yt-dlp not found in path.")
	}

	// Create a temporary directory for the song
	tempDir := filepath.Join("player", "temp")
	if err := os.MkdirAll(tempDir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create temporary directory: %v", err)
	}

	song := bot.pl.track.Name + ".opus"

	// Define the output path
	bot.pl.output = filepath.Join(tempDir, song)

	// Define the yt-dpl command arguments
	args := []string{
		url,
		"--extract-audio",
		"--audio-format", "opus",
		"--output", bot.pl.output,
		"--quiet",
		"--no-playlist",
		"--ignore-errors", // Ignores unavailable videos
		"--no-warnings",
	}

	// Execute the yt-dpl command
	cmd := exec.Command(youtubeDownloader, args...)
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to download and convert video: %v", err)
	}

	return nil
}

func (bot *Bot) cleanUpDir() error {
	tempDir := "./player"
	dir, err := os.Open(tempDir)
	if err != nil {
		return fmt.Errorf("failed to open temp directory: %v", err)
	}
	defer dir.Close()

	files, err := dir.Readdir(0)
	if err != nil {
		return fmt.Errorf("failed to read temp directory: %v", err)
	}

	for _, file := range files {
		err := os.Remove(filepath.Join(tempDir, file.Name()))
		if err != nil {
			fmt.Printf("failed to delete file %s: %v\n", file.Name(), err)
		}
	}

	return nil
}
