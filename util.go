package main

import "os"

func getGuildID() string {
	return os.Getenv("GuildID")
}

func getChannelID() string {
	return os.Getenv("ChannelID")
}

func getBotToken() string {
	return os.Getenv("BotToken")
}
