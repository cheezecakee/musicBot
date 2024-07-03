package bot

import (
	"context"
	"discordBot/app/auth"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/zmb3/spotify/v2"
)

var (
	Clients *auth.Clients
	pl      PlayList
)

type Bot struct {
	Session         *discordgo.Session
	Message         *discordgo.MessageCreate
	Guild           *discordgo.Guild
	State           *discordgo.State
	VoiceConnection *discordgo.VoiceConnection
	VoiceState      *discordgo.VoiceState
	Context         context.Context
	pl              PlayList
}

type PlayList struct {
	trackName string
	track     *spotify.FullTrack
	videoID   string
	output    string
}

func Run() {
	session := Clients.Discord
	context := context.Background()

	// Create a new bot instance
	bot := &Bot{
		Session: session,
		Context: context,
	}

	// Add an event handler
	session.AddHandler(bot.newMessage)

	// Open session
	err := session.Open()
	if err != nil {
		log.Fatalf("Error opening Discord session: %v", err)
	} else {
		log.Println("Discord session successfully started.")
	}
	defer session.Close()

	fmt.Println("Bot running...")
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
}

func (bot *Bot) sendMessage(content string) {
	bot.Session.ChannelMessageSend(bot.Message.ChannelID, content)
}

func (bot *Bot) newMessage(session *discordgo.Session, message *discordgo.MessageCreate) {
	// Initialize the bot instance
	bot.Message = message

	channel, err := session.State.Channel(message.ChannelID)
	if err != nil {
		log.Println("Error getting channel:", err)
		return
	}

	guild, err := session.State.Guild(channel.GuildID)
	if err != nil {
		log.Println("Error getting guild:", err)
		return
	}

	bot.Guild = guild

	// Prevent bot from responding to its own messages
	if bot.Message.Author.ID == bot.Session.State.User.ID {
		return
	}

	// Parse and handle commands
	if err := bot.HandleCommand(); err != nil {
		log.Println("Command error:", err)
	}
}

// Find out who the userID is
func (bot *Bot) getVoiceState() error {
	for _, vs := range bot.Guild.VoiceStates {
		if vs.UserID == bot.Message.Author.ID {
			bot.VoiceState = vs
			return nil
		}
	}
	return fmt.Errorf("user not in a voice channel")
}

func (bot *Bot) getTrackName() {
	trackName := strings.TrimSpace(strings.TrimPrefix(bot.Message.Content, "!play"))
	if trackName == "" {
		bot.sendMessage("Invalid song name.")
		return
	}
	bot.pl.trackName = trackName
}

func (bot *Bot) HandleCommand() error {
	switch {
	case strings.Contains(bot.Message.Content, "!hello"):
		return bot.handleHelloCommand()
	case strings.Contains(bot.Message.Content, "!bye"):
		return bot.handleByeCommand()
	case strings.Contains(bot.Message.Content, "!join"):
		if err := bot.handleJoinCommand(); err != nil {
			return err
		}
	case strings.Contains(bot.Message.Content, "!leave"):
		bot.handleLeaveCommand()
	case strings.Contains(bot.Message.Content, "!play"):
		if err := bot.handlePlayCommand(); err != nil {
			return err
		}
	default:
		return nil
	}
	return nil
}

func (bot *Bot) handleHelloCommand() error {
	bot.sendMessage("Hello World 😃")
	return nil
}

func (bot *Bot) handleByeCommand() error {
	bot.sendMessage("Good bye 👋")
	return nil
}

func (bot *Bot) handleJoinCommand() error {
	// Find the voice state of the user
	err := bot.getVoiceState()
	if err != nil {
		bot.sendMessage("You must be in a voice channel to use this command.")
		return fmt.Errorf("user not in a voice channel")
	}

	// Join the voice channel
	voice, err := bot.Session.ChannelVoiceJoin(bot.Guild.ID, bot.VoiceState.ChannelID, false, false)
	if err != nil {
		bot.sendMessage("Failed to join the voice channel.")
		return fmt.Errorf("failed to join voice channel: %v", err)
	}

	bot.VoiceConnection = voice

	bot.sendMessage("Joined the voice channel!")
	return nil
}

func (bot *Bot) handleLeaveCommand() {
	if bot.VoiceConnection != nil {
		bot.VoiceConnection.Disconnect()
		bot.VoiceConnection = nil
		bot.sendMessage("Left the voice channel!")
	} else {
		bot.sendMessage("I'm not in a voice channel.")
	}
}

func (bot *Bot) handlePlayCommand() error {
	bot.getTrackName()
	err := bot.downloadAudio()
	if err != nil {
		return err
	}
	// Play audio
	err = bot.streamAudio()
	if err != nil {
		bot.sendMessage(fmt.Sprintf("Error: %v", err))
		return err
	}
	// err := bot.testAudio()
	// if err != nil {
	// 	return err
	// }
	// bot.sendMessage(fmt.Sprintf("Now playing: %s by %s", track.Name, track.Artists[0].Name))
	return nil
}
