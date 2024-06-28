package main

import (
	app "discordBot/app"
	bot "discordBot/bot"
	"log"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	app.InitSpotify()

	if err := app.InitYouTube(); err != nil {
		log.Fatal("Error initializing YouTube client:", err)
	}

	bot.BotToken = os.Getenv("DISCORD_BOT_TOKEN")
	bot.Run()
}
