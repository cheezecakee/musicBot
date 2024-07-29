package nlp

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"

	vosk "github.com/alphacep/vosk-api/go"
	"github.com/bwmarrin/discordgo"
	"layeh.com/gopus"
)

type STTResult struct {
	Text string `json:"text"`
}

var (
	model, _    = vosk.NewModel("./nlp/vosk_models/en")
	stt, _      = vosk.NewRecognizer(model, 48000)
	speakers, _ = gopus.NewDecoder(48000, 1)
)

func HandleVoice(c chan *discordgo.Packet, channelID, userID string, Session *discordgo.Session) {
	buffer := new(bytes.Buffer)
	for {
		select {
		case s, ok := <-c:
			if !ok {
				break
			}
			if buffer == nil {
				buffer = new(bytes.Buffer)
			}
			packet, _ := speakers.Decode(s.Opus, 960, false) // frameSize is 960(20ms)
			pcm := new(bytes.Buffer)
			binary.Write(pcm, binary.LittleEndian, packet)
			buffer.Write(pcm.Bytes())
			stt.AcceptWaveform(pcm.Bytes())

			var dur float32 = (float32(len(buffer.Bytes())) / 48000 / 2) // duration of audio

			// When silence packet detected, send result (skip audio shorter than 500ms)
			if dur > 0.5 && len(s.Opus) == 3 && s.Opus[0] == 248 && s.Opus[1] == 255 && s.Opus[2] == 254 {
				log.Println("dur", dur)
				var result STTResult
				json.Unmarshal([]byte(stt.FinalResult()), &result)
				if len(result.Text) > 0 {
					log.Println(fmt.Sprintf("%s: %s", userID, result.Text))
					// process the transiption result:
					Session.ChannelMessageSend(channelID, fmt.Sprintf("%s: %s", userID, result.Text))
				}
				buffer.Reset()
			}
		}
	}
}
