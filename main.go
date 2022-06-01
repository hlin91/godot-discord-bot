package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/harvlin/godot/code"
	"github.com/harvlin/godot/homework"
	"github.com/harvlin/godot/module"
	"github.com/harvlin/godot/rss"
	"github.com/harvlin/godot/voice"
	"golang.org/x/net/html"
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
var componentHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){}
var modalHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){}

func init() {
	flag.Parse()
	modules = append(modules, rss.GetModule())
	modules = append(modules, voice.GetModule())
	modules = append(modules, homework.GetModule())
	modules = append(modules, code.GetModule())
}

func init() {
	rss.AddFeed(`http://fiu758.blog111.fc2.com/?xml`, func(n *html.Node) bool {
		parentFilter := rss.ParentNodeFilterFunc(rss.FilterByClass("main_txt"))
		nodeFilter := rss.DefaultFilterStrategy()
		return parentFilter(n) && nodeFilter(n)
	}, func(n *html.Node) bool {
		parentFilter := rss.ParentNodeFilterFunc(rss.FilterByClass("sh_fc2blogheadbar_body"))
		nodeFilter := rss.DefaultFilterStrategy()
		return parentFilter(n) && nodeFilter(n)
	}, rss.DefaultExtractionStrategy(), rss.DefaultTransformStrategy(), getChannelId, 1)
	rss.AddFeed(`http://2chav.com/?xml`, func(n *html.Node) bool {
		parentFilter := rss.ParentNodeFilterFunc(rss.FilterByClass("kobetu_kiji"))
		nodeFilter := rss.DefaultFilterStrategy()
		return parentFilter(n) && nodeFilter(n)
	}, rss.DefaultFilterStrategy(), rss.DefaultExtractionStrategy(), rss.DefaultTransformStrategy(), getChannelId, 1)
	rss.AddFeed(`https://dlsite-rss.s3-ap-northeast-1.amazonaws.com/voice_rss.xml`, rss.FilterByAttr("property", "og:image"), func(n *html.Node) bool {
		parentFilter := rss.ParentNodeFilterFunc(rss.FilterByClass("logo"))
		nodeFilter := rss.DefaultFilterStrategy()
		return parentFilter(n) && nodeFilter(n)
	}, rss.DefaultExtractionStrategy(), func(s string) string {
		if strings.HasPrefix(s, "/") {
			return `https://www.dlsite.com` + s
		}
		return s
	}, getSecondChannelId, 1)
	rss.AddFeed(`http://avohayo.blog.fc2.com/?xml`, func(n *html.Node) bool {
		parentFilter := rss.ParentNodeFilterFunc(rss.FilterByClass("entry_body"))
		nodeFilter := rss.DefaultFilterStrategy()
		return parentFilter(n) && nodeFilter(n)
	}, func(n *html.Node) bool {
		parentFilter := rss.ParentNodeFilterFunc(rss.FilterByAttr("id", "sh_fc2blogheadbar_menu"))
		nodeFilter := rss.DefaultFilterStrategy()
		return parentFilter(n) && nodeFilter(n)
	}, rss.DefaultExtractionStrategy(), rss.DefaultTransformStrategy(), getChannelId, 1)
}

func main() {
	var session *discordgo.Session
	done := make(chan interface{})
	ret := make(chan interface{})
	session, err := discordgo.New("Bot " + getBotToken())
	if err != nil {
		log.Fatalf("could not create discord session: %v", err)
	}
	ug, err := session.UserGuilds(100, "", "")
	if err != nil {
		log.Panicf("could not retrieve user guilds: %v", err)
	}

	// Starts the rss listener
	session.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Println("starting rss listener...")
		tick := time.NewTicker(WAIT_TIME)
		go rss.ListenerProcess(s, tick, done, ret)
	})

	// Calls the corresponding handler for a command
	session.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		switch i.Type {
		case discordgo.InteractionApplicationCommand: // Handle normal commands
			if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
				h(s, i)
			} else {
				// Unregister the problematic command
				v := i.ApplicationCommandData()
				for _, g := range ug {
					err := session.ApplicationCommandDelete(session.State.User.ID, g.ID, v.ID)
					if err != nil {
						log.Printf("could not delete '%v' command (id %v): %v", v.Name, v.ID, err)
					}
				}
			}
		case discordgo.InteractionMessageComponent: // Handle component commands
			if h, ok := componentHandlers[i.MessageComponentData().CustomID]; ok {
				h(s, i)
			}
		case discordgo.InteractionModalSubmit:
			if h, ok := modalHandlers[i.ModalSubmitData().CustomID]; ok {
				h(s, i)
			}
		}
	})

	session.Identify.Intents = discordgo.IntentsAll
	err = session.Open()
	if err != nil {
		log.Fatalf("could not open bot session: %v", err)
	}
	defer session.Close()
	log.Printf("loading handlers...")
	// Load command modules
	for _, m := range modules {
		m.Load(session, commandHandlers, componentHandlers, modalHandlers, *registerCommands)
	}
	log.Printf("bot is ready")
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop
	log.Println("shutting down...")
	if *removeCommands {
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
