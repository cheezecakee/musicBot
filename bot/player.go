package bot

import (
	"discordBot/app"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/jonas747/dca"
)

func (bot *Bot) convertToDCA() error {
	// Construct the ffmpeg command to convert the audio file to DCA format
	cmd := exec.Command("ffmpeg", "-i", bot.pl.output, "-f", "s16le", "-ac", "2", "-ar", "48000", "-acodec", "pcm_s16le", "-")

	// Pipe the output to dca
	dcaCmd := exec.Command("dca")
	dcaCmd.Stdin, _ = cmd.StdoutPipe()
	dcaCmd.Stdout, _ = os.Create("output.dca")

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

	fmt.Println("Conversion to DCA completed successfully")
	return nil
}

func (bot *Bot) encodeAudio() error {
	encodingSession, err := dca.EncodeFile(bot.pl.output, dca.StdEncodeOptions)
	if err != nil {
		return fmt.Errorf("error encoding file %v", err)
	}
	defer encodingSession.Cleanup()

	output, err := os.Create("output.dca")
	if err != nil {
		// Handle the error
	}
	io.Copy(output, encodingSession)

	return nil
}

func (bot *Bot) streamAudio() error {
	err := bot.handleJoinCommand()
	if err != nil {
		return (err)
	}
	bot.VoiceConnection.Speaking(true)

	bot.convertToDCA()

	file, err := os.Open("output.dca")
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
				return fmt.Errorf("error reading Opus frame: %v", err)
			}
			break // End of audio stream
		}

		// Send the Opus frame to Discord
		select {
		case bot.VoiceConnection.OpusSend <- frame:
		case <-time.After(time.Second):
			return fmt.Errorf("timeout sending Opus frame to Discord")
		}
	}

	fmt.Println("Audio stream completed successfully")
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
	sanitizedTrackName := strings.ReplaceAll(bot.pl.track.Name, " ", "_")
	outputDir := filepath.Join("player", sanitizedTrackName)
	if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create output directory: %v", err)
	}

	// Define the output path
	bot.pl.output = filepath.Join(outputDir, "song.mp3")

	// Define the yt-dpl command arguments
	args := []string{
		url,
		"--extract-audio",
		"--audio-format", "mp3",
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

	fmt.Println(bot.pl.output)

	return nil
}

func (bot *Bot) downloadAudio() error {
	// Search for the track on Spotify
	track, err := app.SearchTrack(bot.pl.trackName)
	if err != nil {
		return fmt.Errorf("error searching track: %v", err)
	}
	bot.pl.track = track
	// Use the track name and artist for the Youtube search query
	query := fmt.Sprintf("%s %s", track.Name, track.Artists[0].Name)
	bot.pl.videoID, err = app.SearchVideo(query)
	if err != nil {
		return fmt.Errorf("error search video: %v", err)
	}

	// Convert the video to audio
	err = bot.convertVideo()
	if err != nil {
		return fmt.Errorf("error converting video: %v", err)
	}

	fmt.Println("video ID:", bot.pl.videoID)
	return nil
}

func (bot *Bot) testAudio() error {
	// Join the voice channel
	err := bot.handleJoinCommand()
	if err != nil {
		return err
	}

	bot.VoiceConnection.Speaking(true)

	file, err := os.Open("test.dca")
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
				return fmt.Errorf("error reading Opus frame: %v", err)
			}
			break // End of audio stream
		}

		// Send the Opus frame to Discord
		select {
		case bot.VoiceConnection.OpusSend <- frame:
		case <-time.After(time.Second):
			return fmt.Errorf("timeout sending Opus frame to Discord")
		}
	}

	fmt.Println("Audio stream completed successfully")
	return nil
}
