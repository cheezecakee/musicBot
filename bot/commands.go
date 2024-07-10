package bot

import (
	"discordBot/player"
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
)

func (bot *Bot) ready(s *discordgo.Session, event *discordgo.Ready) {
	s.UpdateGameStatus(0, "with slash commands!")
}

func (bot *Bot) interactionCreate(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type == discordgo.InteractionApplicationCommand {
		switch i.ApplicationCommandData().Name {
		case "ping":
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Pong!",
				},
			})

		case "play":
			bot.setup(i)
			bot.playCommand(i)
		case "queue":
			bot.sendResponse(i.Interaction, bot.Queue.String())
		case "skip":
			bot.skipCommand(i)
		case "prev":
			bot.Queue.Previous()
			bot.sendResponse(i.Interaction, fmt.Sprintf("Back to: %s", bot.Queue.GetCurrentSong().Name))
		case "remove":
			index := int(i.ApplicationCommandData().Options[0].IntValue())
			bot.Queue.RemoveSong(index)
			bot.sendResponse(i.Interaction, fmt.Sprintf("Removed song at index %d", index))
		}
	}
}

func (bot *Bot) registerCommands(s *discordgo.Session) {
	commands := []*discordgo.ApplicationCommand{
		{
			Name:        "ping",
			Description: "Replies with Pong!",
		},
		{
			Name:        "play",
			Description: "Play a song",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "song",
					Description: "The name of the song to play",
					Type:        discordgo.ApplicationCommandOptionString,
					Required:    true,
				},
			},
		},
		{
			Name:        "queue",
			Description: "Show the current queue",
		},
		{
			Name:        "skip",
			Description: "Skip to the next song in the queue",
		},
		{
			Name:        "prev",
			Description: "Go back to the previous song in the queue",
		},
		{
			Name:        "remove",
			Description: "Remove a song from the queue",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "index",
					Description: "The index of the song to remove",
					Type:        discordgo.ApplicationCommandOptionInteger,
					Required:    true,
				},
			},
		},
	}

	for _, v := range commands {
		_, err := s.ApplicationCommandCreate(s.State.User.ID, "", v)
		if err != nil {
			fmt.Printf("Cannot create '%v' command: %v\n", v.Name, err)
		}
	}
}

func (bot *Bot) unregisterCommands(s *discordgo.Session) {
	commands, err := s.ApplicationCommands(s.State.User.ID, "")
	if err != nil {
		fmt.Printf("Could not fetch registered commands: %v", err)
		return
	}

	for _, v := range commands {
		err := s.ApplicationCommandDelete(s.State.User.ID, "", v.ID)
		if err != nil {
			fmt.Printf("Cannot delete '%v' command: %v\n", v.Name, err)
		}
	}
}

func (bot *Bot) sendResponse(i *discordgo.Interaction, response string) {
	err := bot.Session.InteractionRespond(i, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: response,
		},
	})
	if err != nil {
		fmt.Printf("Cannot send response: %v", err)
	}
}

func (bot *Bot) playCommand(i *discordgo.InteractionCreate) {
	userID := i.Member.User.ID

	// Send the now playing or added to queue message
	if len(bot.Queue.Songs) == 1 {
		bot.sendResponse(i.Interaction, fmt.Sprintf("Now playing: %s by %s", bot.Player.Track.Name, bot.Player.Track.Artists[0].Name))
	} else {
		bot.sendResponse(i.Interaction, fmt.Sprintf("Added to queue: %s by %s", bot.Player.Track.Name, bot.Player.Track.Artists[0].Name))
	}

	if err := bot.handleJoinCommand(userID); err != nil {
		bot.sendResponse(i.Interaction, fmt.Sprintf("Error: %v", err))
		return
	}

	// Start streaming if not already playing
	if len(bot.Queue.Songs) == 1 {
		bot.Player.Play(bot.VoiceConnection, &bot.Queue)
	}
}

func (bot *Bot) setup(i *discordgo.InteractionCreate) {
	trackName := i.ApplicationCommandData().Options[0].StringValue()

	// Acknowledge the interaction
	err := bot.Session.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Processing your request...",
		},
	})
	if err != nil {
		log.Printf("Cannot send response: %v", err)
		return
	}

	// Set the track name
	bot.Player.Name = trackName
	if err := bot.Player.DownloadAudio(); err != nil {
		bot.sendResponse(i.Interaction, fmt.Sprintf("Error: %v", err))
		return
	}

	// Download Audio DCA file
	err = bot.Player.ConvertToDCA()
	if err != nil {
		fmt.Printf("Error converting video: %v\n", err)
		return
	}

	// Add song to queue with DcaPath
	bot.Queue.AddSong(player.Song{Name: bot.Player.Track.Name, Artist: bot.Player.Track.Artists[0].Name, DcaPath: bot.Player.DcaPath})
}

func (bot *Bot) skipCommand(i *discordgo.InteractionCreate) {
	// Acknowledge the interaction immediately to avoid multiple responses
	err := bot.Session.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Skipping the current song...",
		},
	})
	if err != nil {
		log.Printf("Cannot send response: %v", err)
		return
	}

	bot.Queue.Next()
	// Send skip signal
	bot.Player.Skip <- true
}
