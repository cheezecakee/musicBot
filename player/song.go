package player

import (
	"discordBot/app"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func (p *Player) ConvertToDCA() error {
	// Construct the ffmpeg command to convert the audio file to DCA format
	cmd := exec.Command("ffmpeg", "-i", p.OpusPath, "-f", "s16le", "-ac", "2", "-ar", "48000", "-acodec", "pcm_s16le", "-")

	// Create path with song name
	p.DcaPath = "./player/temp/" + p.Track.Name + ".dca"

	// Pipe the output to dca
	dcaCmd := exec.Command("dca")
	dcaCmd.Stdin, _ = cmd.StdoutPipe()
	dcaCmd.Stdout, _ = os.Create(p.DcaPath)

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

	// fmt.Println(p.DcaPath)

	// fmt.Println("Conversion to DCA completed successfully")
	return nil
}

func (p *Player) DownloadAudio() error {
	// Search for the track on Spotify
	track, err := app.SearchTrack(p.Name)
	if err != nil {
		return fmt.Errorf("error searching track: %v", err)
	}
	p.Track = track
	// Use the track name and artist for the Youtube search query
	query := fmt.Sprintf("%s %s official audio lyric", track.Name, track.Artists[0].Name)
	p.VideoID, err = app.SearchVideo(query)
	if err != nil {
		return fmt.Errorf("error search video: %v", err)
	}

	// Convert the video to audio
	err = p.convertVideo()
	if err != nil {
		return fmt.Errorf("error converting video: %v", err)
	}

	// fmt.Println("video ID:", bot.pl.videoID)
	return nil
}

func (p *Player) convertVideo() error {
	// Construst the download URL
	url := fmt.Sprintf("https://www.youtube.com/watch?v=%s", p.VideoID)

	youtubeDownloader, err := exec.LookPath("yt-dlp")
	if err != nil {
		fmt.Println("yt-dlp not found in path.")
	}

	// Create a temporary directory for the song
	tempDir := filepath.Join("player", "temp")
	if err := os.MkdirAll(tempDir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create temporary directory: %v", err)
	}

	song := p.Track.Name + ".opus"

	// Define the output path
	p.OpusPath = filepath.Join(tempDir, song)

	// Define the yt-dpl command arguments
	args := []string{
		url,
		"--extract-audio",
		"--audio-format", "opus",
		"--output", p.OpusPath,
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

// Delete songs after it's done playing
func cleanUpDir() error {
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
