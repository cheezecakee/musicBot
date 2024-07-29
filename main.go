package main

import (
	"discordBot/app"
	"discordBot/app/auth"
	"discordBot/bot"
	"log"
)

func main() {
	tokens, err := auth.LoadTokens()
	if err != nil {
		log.Fatalf("Error loading tokens: %v", err)
	}

	clients, err := auth.InitClients(tokens)
	if err != nil {
		log.Fatalf("Error initializing clients: %v", err)
	}
	// log.Println(clients)

	bot.Clients = clients
	app.Clients = clients
	bot.Run()
}
