package player

import (
	"discordBot/app"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/cheezecakee/dca"
	ytdl "github.com/kkdai/youtube/v2"
)

// Frame represents a single audio frame
type Frame struct {
	data     []byte
	metaData bool
}

// EncodeSession represents the session for encoding and streaming
type EncodeSession struct {
	frameChannel chan *Frame
}

func (p *Player) find() {
	// Search for the track on Spotify
	track, err := app.SearchTrack(p.Name)
	if err != nil {
		fmt.Printf("error searching track: %v\n", err)
	}
	p.Track = track

	// Use the track name and artist for the Youtube search query
	query := fmt.Sprintf("%s %s official audio lyric", track.Name, track.Artists[0].Name)
	p.VideoID, err = app.SearchVideo(query)
	if err != nil {
		fmt.Printf("error searching video: %v\n", err)
	}
}

func (p *Player) DCA(vc *discordgo.VoiceConnection) {
	// Encode audio from the URL
	encodeSession, err := dca.EncodeFile(p.url(), dca.StdEncodeOptions)
	if err != nil {
		log.Fatal("Failed creating an encoding session: ", err)
	}
	log.Println("encodeSession opts:", encodeSession.Options())

	log.Println("encodeSession")
	fmt.Printf("\n%+v\n", encodeSession)

	done := make(chan error)
	stream := dca.NewStream(encodeSession, vc, done)

	ticker := time.NewTicker(time.Second)

	for {
		select {
		case err := <-done:
			if err != nil && err != io.EOF {
				log.Fatal("An error occured", err)
			}

			// Clean up incase something happened and ffmpeg is still running
			encodeSession.Truncate()
			return
		case <-ticker.C:
			stream.PlaybackPosition()
		}
	}
}

// Method to get the direct audio stream URL from YouTube using yt-dlp and stream it to Discord
func (p *Player) url() string {
	p.find()
	// Create a new YouTube client
	client := ytdl.Client{}

	// Get video information from YouTube
	video, err := client.GetVideo(p.VideoID)
	if err != nil {
		log.Printf("Error getting video info: %v\n", err)
	}
	log.Printf("Video Name: %v\n", video.Title)
	log.Printf("Video Duration: %v\n", video.Duration)
	log.Printf("Video ID: %v\n", video.ID)

	formats := video.Formats.WithAudioChannels().Itag(140)
	streamURL, err := client.GetStreamURL(video, &formats[0])
	if err != nil {
		panic(err)
	}
	log.Println("formats")
	log.Printf("%+v\n\n", formats)

	return streamURL
}
