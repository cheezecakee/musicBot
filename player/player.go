package player

import (
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/jonas747/dca"
	"github.com/zmb3/spotify/v2"
)

type Player struct {
	Name     string
	Track    *spotify.FullTrack
	VideoID  string
	OpusPath string
	DcaPath  string
	Skip     chan bool
}

func NewPlayer() *Player {
	return &Player{
		Skip: make(chan bool),
	}
}

func (p *Player) Play(vc *discordgo.VoiceConnection, q *Queue) error {
	for {
		currentSong := q.GetCurrentSong()
		if currentSong == nil {
			return nil
		}
		log.Println(currentSong)

		dcaPath := currentSong.DcaPath
		file, err := os.Open(dcaPath)
		if err != nil {
			fmt.Println("Error opening dca file:", err)
			return err
		}
		defer file.Close()

		decoder := dca.NewDecoder(file)
		vc.Speaking(true)
		defer vc.Speaking(false)

		p.stream(vc, decoder)

		// After finishing current song or skip signal, move to the next song
		q.Next()
		log.Println("Moving to next song")
	}
}

func (p *Player) stream(vc *discordgo.VoiceConnection, decoder *dca.Decoder) error {
	for {
		frame, err := decoder.OpusFrame()
		if err != nil {
			if err != io.EOF {
				return err
			}
			break
		}
		select {
		case vc.OpusSend <- frame:
			// Send the audio frame to the voice connection
		case <-p.Skip:
			log.Println("Skip signal received, skipping to next song")
			return nil // Break the inner loop to continue with the next song
		case <-time.After(time.Second):
			// Timeout case if no action occurs within 1 second
			return nil
		}
	}
	return nil
}
