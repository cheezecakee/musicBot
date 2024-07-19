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
	"github.com/zmb3/spotify/v2"
)

type Player struct {
	Name    string
	Track   *spotify.FullTrack
	VideoID string
	Url     string
	Skip    chan bool
	Prev    chan bool
	Resume  chan bool
	Pause   chan bool
	Stop    chan bool
}

func NewPlayer() *Player {
	return &Player{
		Skip:   make(chan bool),
		Prev:   make(chan bool),
		Pause:  make(chan bool),
		Resume: make(chan bool),
		Stop:   make(chan bool),
	}
}

func (p *Player) Stream(vc *discordgo.VoiceConnection, q *Queue) {
	ticker := time.NewTicker(time.Second)

playLoop:
	for {
		currentSong := q.GetCurrentSong()
		if currentSong == nil {
			log.Println("No more songs in the queue.")
			return
		}

		url := p.url(currentSong.Url)
		log.Println("current song: ", currentSong.Name)
		log.Println("url: ", url)

		vc.Speaking(true)
		defer vc.Speaking(false)

		// Encode audio from the URL
		encodeSession, err := dca.EncodeFile(url, dca.StdEncodeOptions)
		if err != nil {
			log.Fatal("Failed creating an encoding session: ", err)
		}

		done := make(chan error)
		stream := dca.NewStream(encodeSession, vc, done)

		for {
			select {
			case err := <-done:
				if err != nil && err != io.EOF {
					log.Fatal("An error occurred:", err)
				}
				encodeSession.Cleanup()
				log.Println("Song ended, moving to the next song")
				q.Next()
				continue playLoop
			case <-ticker.C:
				stream.PlaybackPosition()
			case <-p.Skip:
				encodeSession.Cleanup()
				log.Println("Skip signal received, skipping to next song")
				q.Next()
				continue playLoop
			case <-p.Prev:
				encodeSession.Cleanup()
				log.Println("Prev signal received, playing previous song")
				q.Previous()
				continue playLoop
			case <-p.Pause:
				stream.SetPaused(true)
				log.Println("Pause signal received, playing previous song")
			case <-p.Resume:
				stream.SetPaused(false)
				log.Println("Play signal received, playing previous song")
			case <-p.Stop:
				encodeSession.Cleanup()
				log.Println("Stop signal received, stopping song and clearing queue")
				q.Clear()
				vc.Disconnect()
				return
			}
		}
	}
}

// Finds the song
func (p *Player) Find() {
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

// Method to get the direct audio stream URL from YouTube using yt-dlp and stream it to Discord
func (p *Player) url(videoID string) string {
	// Create a new YouTube client
	client := ytdl.Client{}

	// Get video information from YouTube
	video, err := client.GetVideo(videoID)
	if err != nil {
		log.Printf("Error getting video info: %v\n", err)
	}
	// log.Printf("Video Name: %v\n", video.Title)
	// log.Printf("Video Duration: %v\n", video.Duration)

	formats := video.Formats.WithAudioChannels().Itag(140)
	streamURL, err := client.GetStreamURL(video, &formats[0])
	if err != nil {
		panic(err)
	}
	// log.Printf("%+v\n\n", formats)

	return streamURL
}
