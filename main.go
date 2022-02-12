package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/harvlin/godot/module"
	"github.com/harvlin/godot/rss"
	"github.com/harvlin/godot/voice"
)

const (
	WAIT_TIME = 20 * time.Minute
)

// Parameters
var (
	registerCommands = flag.Bool("rgcmd", false, "Registers slash commands on start")
	removeCommands   = flag.Bool("rmcmd", false, "Remove all commands after shutdowning or not")
)

var modules = []module.Module{}
var commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){}

func init() {
	flag.Parse()
	modules = append(modules, rss.GetModule())
	modules = append(modules, voice.GetModule())
}

func init() {
	rss.AddFeed(`http://fiu758.blog111.fc2.com/?xml`, "main_txt", "sh_fc2blogheadbar_body", 2)
	rss.AddFeed(`http://2chav.com/?xml`, "kobetu_kiji", "", 2)
}

func main() {
	var session *discordgo.Session
	done := make(chan interface{})
	ret := make(chan interface{})
	session, err := discordgo.New("Bot " + getBotToken())
	if err != nil {
		log.Fatalf("could not create discord session: %v", err)
	}

	// Starts the rss listener
	session.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Println("starting rss listener...")
		tick := time.NewTicker(WAIT_TIME)
		go rss.ListenerProcess(s, getChannelID(), tick, done, ret)
		log.Println("bot is running...")
	})
	// Calls the corresponding handler for a command
	session.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
			h(s, i)
		}
	})

	session.Identify.Intents = discordgo.IntentsAll
	err = session.Open()
	if err != nil {
		log.Fatalf("could not open bot session: %v", err)
	}
	defer session.Close()

	// Load command modules
	for _, m := range modules {
		m.Load(session, commandHandlers, *registerCommands)
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop
	log.Println("shutting down...")
	if *removeCommands {
		ug, err := session.UserGuilds(100, "", "")
		if err != nil {
			log.Panicf("could not retrieve user guilds: %v", err)
		}
		for _, g := range ug {
			cmds, _ := session.ApplicationCommands(session.State.User.ID, g.ID)
			for _, v := range cmds {
				err := session.ApplicationCommandDelete(session.State.User.ID, g.ID, v.ID)
				if err != nil {
					log.Printf("could not delete '%v' command (id %v): %v", v.Name, v.ID, err)
				}
			}
		}
	}
	var signal interface{}
	done <- signal
	<-ret
}
