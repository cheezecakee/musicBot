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
	Queue   *Queue
	Song    *Song
}

func NewPlayer() *Player {
	return &Player{
		Skip:   make(chan bool),
		Prev:   make(chan bool),
		Pause:  make(chan bool),
		Resume: make(chan bool),
		Stop:   make(chan bool),
		Queue:  &Queue{},
		Song:   &Song{},
	}
}

func (p *Player) Stream(vc *discordgo.VoiceConnection) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

playLoop:
	for {
		currentSong := p.Queue.GetCurrentSong()
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
				p.Queue.Next()
				continue playLoop
			case <-ticker.C:
				stream.PlaybackPosition()
			case <-p.Skip:
				encodeSession.Cleanup()
				log.Println("Skip signal received, skipping to next song")
				p.Queue.Next()
				continue playLoop
			case <-p.Prev:
				encodeSession.Cleanup()
				log.Println("Prev signal received, playing previous song")
				p.Queue.Previous()
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
				p.Queue.Clear()
				vc.Disconnect()
				return
			}
		}
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

// Finds the song based on name or spotify link
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

	log.Println("Adding song to queue")
	p.Queue.AddSong(Song{Name: p.Track.Name, Artist: p.Track.Artists[0].Name, Url: p.VideoID})
}

// Finds the song based on spotify playlist link
func (p *Player) FindPlaylistSpotify(playlistID string) {
	// Fetch the playlist details
	tracks, err := app.SearchSpotifyPlaylist(playlistID)
	if err != nil {
		fmt.Printf("error fetching playlist: %v\n", err)
		return
	}

	// Add each track to the queue
	for _, track := range tracks {
		query := fmt.Sprintf("%s %s official audio lyric", track.Name, track.Artists[0].Name)
		videoID, err := app.SearchVideo(query)
		if err != nil {
			fmt.Printf("error searching video: %v\n", err)
			continue
		}
		p.Queue.AddSong(Song{Name: track.Name, Artist: track.Artists[0].Name, Url: videoID})
	}
}

// Finds the song based on youtube song link
func (p *Player) FindYoutube(videoID string) string {
	// Fetch video details
	trackName, err := app.GetTrackInfoFromYouTubeURL(videoID)
	if err != nil {
		fmt.Printf("error fetching track info: %v\n", err)
		return ""
	}
	p.Queue.AddSong(Song{Name: trackName, Url: videoID})

	return trackName
}
