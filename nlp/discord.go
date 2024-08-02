package nlp

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	vosk "github.com/alphacep/vosk-api/go"
	"github.com/bwmarrin/discordgo"
	"layeh.com/gopus"
)

type STTResult struct {
	Text string `json:"text"`
}

type KeyWords struct {
	Play   string
	Resume string
	Pause  string
	Back   string
	Skip   string
	Queue  string
	Remove string
}

var (
	model, _    = vosk.NewModel("./nlp/vosk_models/en")
	stt, _      = vosk.NewRecognizer(model, 48000)
	speakers, _ = gopus.NewDecoder(48000, 1)
	Result      = make(chan string)
	Command     string
	Arg         string
)

func NewKeyWords() *KeyWords {
	return &KeyWords{
		Play:   "play",
		Pause:  "pause",
		Resume: "resume",
		Back:   "back",
		Skip:   "skip",
		Queue:  "q",
		Remove: "remove",
	}
}

func HandleVoice(c chan *discordgo.Packet, channelID, userID string, Session *discordgo.Session) {
	buffer := new(bytes.Buffer)
	for {
		select {
		case s, ok := <-c:
			if !ok {
				log.Println("Not ok", ok)
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

			var duration float32 = (float32(len(buffer.Bytes())) / 48000 / 2) // duration of audio

			// When silence packet detected, send result (skip audio shorter than 500ms)
			if duration > 0.5 && len(s.Opus) == 3 && s.Opus[0] == 248 && s.Opus[1] == 255 && s.Opus[2] == 254 {
				log.Println("dur", duration)
				var result STTResult
				json.Unmarshal([]byte(stt.FinalResult()), &result)

				if len(result.Text) > 0 {
					log.Println(fmt.Sprintf("%s: %s", userID, result.Text))

					command, arg := ParseCommand(result.Text)
					Command = command
					Arg = arg

					// process the transiption result:
					// Session.ChannelMessageSend(channelID, fmt.Sprintf("%s: %s", userID, result.Text))
				}

				buffer.Reset()
			}
		}
	}
}

func ParseCommand(result string) (string, string) {
	words := strings.Fields(result)

	command := strings.ToLower(words[0])
	argument := strings.Join(words[1:], " ")

	return command, argument
}
