package main

import "os"

func getChannelId() string {
	return os.Getenv("ChannelID")
}

func getBotToken() string {
	return os.Getenv("BotToken")
}

func getSecondChannelId() string {
	return os.Getenv("SecondChannelID")
}
