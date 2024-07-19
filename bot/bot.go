package bot

import (
	"discordBot/app/auth"
	"discordBot/player"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"

	"github.com/bwmarrin/discordgo"
)

var Clients *auth.Clients

type Bot struct {
	Clients         *auth.Clients
	Session         *discordgo.Session
	Message         *discordgo.MessageCreate
	Guild           *discordgo.Guild
	State           *discordgo.State
	VoiceConnection *discordgo.VoiceConnection
	VoiceState      *discordgo.VoiceState
	Player          player.Player
	Queue           player.Queue
	Song            player.Song
}

func Run() {
	session := Clients.Discord

	// Create a new bot instance
	bot := &Bot{
		Session: session,
		Player: player.Player{
			Skip: make(chan bool),
			Prev: make(chan bool),
		},
		Queue: player.Queue{},
		Song:  player.Song{},
	}

	// Add an event handler
	session.AddHandler(bot.newMessage)
	session.AddHandler(bot.ready)
	session.AddHandler(bot.interactionCreate)

	// Open session
	err := session.Open()
	if err != nil {
		log.Fatalf("Error opening Discord session: %v", err)
	} else {
		log.Println("Discord session successfully started.")
	}
	defer session.Close()

	bot.registerCommands(session)

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

func (bot *Bot) getVoiceState(userID string) error {
	for _, vs := range bot.Guild.VoiceStates {
		if vs.UserID == userID {
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
	bot.Player.Name = trackName
}

func (bot *Bot) HandleCommand() error {
	switch {
	case strings.Contains(bot.Message.Content, "!hello"):
		return bot.handleHelloCommand()
	case strings.Contains(bot.Message.Content, "!bye"):
		return bot.handleByeCommand()
	case strings.Contains(bot.Message.Content, "!join"):
		if err := bot.handleJoinCommand(bot.Message.Author.ID); err != nil {
			return err
		}
	default:
		return nil
	}
	return nil
}

func (bot *Bot) handleHelloCommand() error {
	bot.sendMessage("Hello! use the / command for more!😃")
	return nil
}

func (bot *Bot) handleByeCommand() error {
	bot.sendMessage("Good bye 👋")
	return nil
}

func (bot *Bot) handleJoinCommand(userID string) error {
	// Find the voice state of the user
	err := bot.getVoiceState(userID)
	if err != nil {
		return fmt.Errorf("user not in a voice channel")
	}

	// Join the voice channel
	voice, err := bot.Session.ChannelVoiceJoin(bot.Guild.ID, bot.VoiceState.ChannelID, false, false)
	if err != nil {
		return fmt.Errorf("failed to join voice channel: %v", err)
	}

	bot.VoiceConnection = voice
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
