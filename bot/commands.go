package bot

import (
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/bwmarrin/discordgo"
)

func (bot *Bot) ready(s *discordgo.Session, event *discordgo.Ready) {
	s.UpdateGameStatus(0, "with slash commands!")
}

func (bot *Bot) interactionCreate(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type == discordgo.InteractionApplicationCommand {
		switch i.ApplicationCommandData().Name {
		case "play":
			bot.sendResponse(i.Interaction, "Processing...")
			bot.setup(i)
			bot.playCommand(i)
		case "queue":
			bot.queueCommand(i.Interaction, bot.Player.Queue.String())
		case "stop":
			bot.sendResponse(i.Interaction, "Processing...")
			bot.stopCommand(i)
		case "pause":
			bot.sendResponse(i.Interaction, "Processing...")
			bot.pauseCommand(i)
		case "resume":
			bot.sendResponse(i.Interaction, "Processing...")
			bot.resumeCommand(i)
		case "skip":
			bot.sendResponse(i.Interaction, "Processing...")
			bot.skipCommand(i)
		case "prev":
			bot.sendResponse(i.Interaction, "Processing...")
			bot.prevCommand(i)
		case "remove":
			bot.sendResponse(i.Interaction, "Processing...")
			bot.removeCommand(i)
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
			Name:        "stop",
			Description: "Stops song and disconnects musicBot",
		},
		{
			Name:        "pause",
			Description: "Pause song",
		},
		{
			Name:        "resume",
			Description: "Resume song",
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
			Name:        "clear",
			Description: "Clear queue",
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

	var wg sync.WaitGroup

	for _, cmd := range commands {
		wg.Add(1)
		go func(command *discordgo.ApplicationCommand) {
			defer wg.Done()
			// log.Printf("Registering command: %v", command.Name) // Add this line for debugging
			_, err := s.ApplicationCommandCreate(s.State.User.ID, "", command)
			if err != nil {
				fmt.Printf("Cannot create '%v' command: %v\n", command.Name, err)
			}
		}(cmd)
	}

	wg.Wait()
	log.Println("All commands registered")
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
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: response,
		},
	})
	if err != nil {
		fmt.Printf("Cannot send response: %v", err)
	}
}

func (bot *Bot) sendFollowUp(i *discordgo.InteractionCreate, response string) {
	_, err := bot.Session.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
		Content: response,
	})
	if err != nil {
		log.Printf("Cannot send follow-up response: %v", err)
	}
}

func (bot *Bot) queueCommand(i *discordgo.Interaction, response string) {
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

func (bot *Bot) removeCommand(i *discordgo.InteractionCreate) {
	index := int(i.ApplicationCommandData().Options[0].IntValue())

	song := bot.Player.Queue.Songs[index-1]

	bot.Player.Queue.RemoveSong(index - 1)
	bot.sendFollowUp(i, fmt.Sprintf("Removed song %s by %s", song.Name, song.Artist))
}

func (bot *Bot) clearCommand(i *discordgo.InteractionCreate) {
	bot.Player.Queue.Clear()
	bot.sendFollowUp(i, "Queue has been cleared")
}

func (bot *Bot) stopCommand(i *discordgo.InteractionCreate) {
	go func() {
		bot.Player.Stop <- true
	}()

	bot.sendFollowUp(i, "Leaving the voice channel")
}

func (bot *Bot) pauseCommand(i *discordgo.InteractionCreate) {
	go func() {
		bot.Player.Pause <- true
	}()

	bot.sendFollowUp(i, "Song Paused")
}

func (bot *Bot) resumeCommand(i *discordgo.InteractionCreate) {
	go func() {
		bot.Player.Resume <- true
	}()

	bot.sendFollowUp(i, "Resuming Song")
}

func (bot *Bot) skipCommand(i *discordgo.InteractionCreate) {
	go func() {
		bot.Player.Skip <- true
	}()

	bot.sendFollowUp(i, fmt.Sprintf("Song skipped to: %s by %s", bot.Player.Track.Name, bot.Player.Track.Artists[0].Name))
}

func (bot *Bot) prevCommand(i *discordgo.InteractionCreate) {
	go func() {
		bot.Player.Prev <- true
	}()
	bot.sendFollowUp(i, fmt.Sprintf("Going back to the previous song: %s by %s", bot.Player.Track.Name, bot.Player.Track.Artists[0].Name))
}

func (bot *Bot) playCommand(i *discordgo.InteractionCreate) {
	userID := i.Member.User.ID

	if err := bot.handleJoinCommand(userID); err != nil {
		log.Printf("Error in handleJoinCommand: %v", err)
		return
	}

	// Start streaming if not already playing
	if len(bot.Player.Queue.Songs) == 1 {
		bot.Player.Stream(bot.VoiceConnection)
	}
}

func (bot *Bot) setup(i *discordgo.InteractionCreate) {
	trackName := i.ApplicationCommandData().Options[0].StringValue()
	log.Println("setup invoked with input:", trackName)

	// Find the appropriate handler based on the URL prefix
	for prefix, handler := range urlHandlers {
		if strings.HasPrefix(trackName, prefix) {
			handler(bot, trackName, i)
			return
		}
	}

	// Default handling for unknown URLs
	bot.Player.Name = trackName
	bot.Player.Find()
	bot.sendFollowUp(i, fmt.Sprintf("Added to queue: %s by %s", bot.Player.Track.Name, bot.Player.Track.Artists[0].Name))
}
