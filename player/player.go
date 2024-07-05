package player

import (
	"fmt"
	"io"
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
}

func (p *Player) StreamAudio(vc *discordgo.VoiceConnection, q *Queue) error {
	for q.Current < len(q.Songs) {
		// Prepare the current song
		currentSong := q.CurrentSong()
		if currentSong == nil {
			break
		}

		// Set the DCA path based on the current song
		p.DcaPath = fmt.Sprintf("%s.dca", currentSong.Name)

		// Ensure the DCA file exists
		err := p.convertToDCA()
		if err != nil {
			fmt.Printf("Error converting video: %v\n", err)
			return err
		}

		file, err := os.Open(p.DcaPath)
		if err != nil {
			fmt.Println("Error opening dca file:", err)
			return err
		}

		// Create a DCA decoder to read from the encoded DCA file
		decoder := dca.NewDecoder(file)
		vc.Speaking(true)

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
			case vc.OpusSend <- frame:
			case <-time.After(time.Second):
				return nil
			}
		}

		vc.Speaking(false)
		file.Close()
		q.Next() // Move to the next song in the queue
	}

	vc.Disconnect()
	fmt.Println("Audio stream completed successfully")
	return nil
}
