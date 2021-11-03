package main

import "os"

func getChannelID() string {
	return os.Getenv("ChannelID")
}

func getBotToken() string {
	return os.Getenv("BotToken")
}
