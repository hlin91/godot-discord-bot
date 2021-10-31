package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/harvlin/godot/rss"
	"github.com/mmcdole/gofeed"
)

const (
	WAIT_TIME = 1 * time.Minute
)

// Parameters
var (
	RemoveCommands = flag.Bool("rmcmd", true, "Remove all commands after shutdowning or not")
)

func init() {
	flag.Parse()
}

func init() {
	rss.AddFeed(`http://fiu758.blog111.fc2.com/?xml`, "main_txt", 2)
	rss.AddFeed(`http://2chav.com/?xml`, "kobetu_kiji", 2)
}

// Declare our slash commands and their handlers
var (
	commands = []*discordgo.ApplicationCommand{
		{
			Name:        "pinger",
			Description: "Replies with ponger",
		},
		{
			Name:        "test_rss",
			Description: "Test the RSS feed feature by posting an RSS post to the channel",
		},
	}
	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"pinger": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{Content: "ponger"},
			})
		},
		"test_rss": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "",
				},
			})
			if err != nil {
				s.FollowupMessageCreate(s.State.User.ID, i.Interaction, true, &discordgo.WebhookParams{
					Content: "Something went wrong",
				})
				return
			}
			// Do RSS stuff here
			rss.ClearHistory()
			items := rss.GetLatest()
			var item *gofeed.Item
			images := []string{}
			for key, val := range items {
				item = val[0]
				images, _ = rss.GetImages(item.Link, key.Class, key.NumImages)
				break
			}
			s.InteractionResponseEdit(s.State.User.ID, i.Interaction, &discordgo.WebhookEdit{
				Embeds: []*discordgo.MessageEmbed{
					rss.ItemToEmbed(item, images),
				},
			})
		},
	}
)

func main() {
	var session *discordgo.Session
	done := make(chan interface{})
	ret := make(chan interface{})
	session, err := discordgo.New("Bot " + getBotToken())
	if err != nil {
		log.Fatalf("could not create discord session: %v", err)
	}
	session.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Println("starting rss listener...")
		tick := time.NewTicker(WAIT_TIME)
		go rss.ListenerProcess(s, getChannelID(), tick, done, ret)
	})
	session.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
			h(s, i)
		}
	})
	err = session.Open()
	if err != nil {
		log.Fatalf("could not open bot session: %v", err)
	}
	defer session.Close()
	for _, v := range commands {
		_, err := session.ApplicationCommandCreate(session.State.User.ID, getGuildID(), v)
		if err != nil {
			log.Panicf("could not create '%v' command: %v", v.Name, err)
		}
	}
	session.Identify.Intents = discordgo.IntentsAll
	stop := make(chan os.Signal)
	signal.Notify(stop, os.Interrupt)
	<-stop
	log.Println("shutting down...")
	if *RemoveCommands {
		cmds, _ := session.ApplicationCommands(session.State.User.ID, getGuildID())
		for _, v := range cmds {
			err := session.ApplicationCommandDelete(session.State.User.ID, getGuildID(), v.ID)
			if err != nil {
				log.Panicf("could not delete '%v' command (id %v): %v", v.Name, v.ID, err)
			}
		}
	}
	var signal interface{}
	done <- signal
	<-ret
}
