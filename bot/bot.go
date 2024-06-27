package bot

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"

	"github.com/bwmarrin/discordgo"
)

var BotToken string

func checkNilErr(e error) {
	if e != nil {
		log.Fatal(e)
	}
}

func Run() {
	// create a session
	discord, err := discordgo.New("Bot " + BotToken)
	checkNilErr(err)

	// add an event handler
	discord.AddHandler(newMessage)

	// open session
	err = discord.Open()
	if err != nil {
		log.Fatal(err)
	}
	defer discord.Close()

	// keep bot running until there is an OS interruption (ctrl + C)
	fmt.Println("Bot running...")
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
}

func newMessage(discord *discordgo.Session, message *discordgo.MessageCreate) {
	// prevent bot from responding to its own messages
	if message.Author.ID == discord.State.User.ID {
		return
	}

	// get the channel information
	channel, err := discord.State.Channel(message.ChannelID)
	if err != nil {
		log.Println("Error getting channel:", err)
		return
	}

	// get the guild information
	guild, err := discord.State.Guild(channel.GuildID)
	if err != nil {
		log.Println("Error getting guild:", err)
		return
	}

	// respond to user message if it contains `!help`, `!bye`, or `!join`
	switch {
	case strings.Contains(message.Content, "!hello"):
		discord.ChannelMessageSend(message.ChannelID, "Hello World😃")
	case strings.Contains(message.Content, "!bye"):
		discord.ChannelMessageSend(message.ChannelID, "Good bye👋")
	case strings.Contains(message.Content, "!join"):
		// find the voice state of the user
		var voiceState *discordgo.VoiceState
		for _, vs := range guild.VoiceStates {
			if vs.UserID == message.Author.ID {
				voiceState = vs
				break
			}
		}

		if voiceState == nil {
			discord.ChannelMessageSend(message.ChannelID, "You must be in a voice channel to use this command.")
			return
		}

		// join the voice channel
		voice, err := discord.ChannelVoiceJoin(guild.ID, voiceState.ChannelID, false, false)
		if err != nil {
			discord.ChannelMessageSend(message.ChannelID, "Failed to join the voice channel.")
			log.Println("Error joining the voice channel:", err)
			return
		}

		go func() {
			if strings.Contains(message.Content, "!leave") {
				voice.Disconnect()
				discord.ChannelMessageSend(message.ChannelID, "Left the voice channel.")
				return
			}
		}()

		discord.ChannelMessageSend(message.ChannelID, "Joined the voice channel!")
	}
}
