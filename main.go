package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/harvlin/godot/rss"
	"github.com/harvlin/godot/voice"
	"github.com/mmcdole/gofeed"
)

const (
	WAIT_TIME = 20 * time.Minute
)

// Parameters
var (
	RemoveCommands = flag.Bool("rmcmd", false, "Remove all commands after shutdowning or not")
)

func init() {
	flag.Parse()
}

func init() {
	rss.AddFeed(`http://fiu758.blog111.fc2.com/?xml`, "main_txt", "sh_fc2blogheadbar_body", 2)
	rss.AddFeed(`http://2chav.com/?xml`, "kobetu_kiji", "", 2)
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
		{
			Name:        "join",
			Description: "Just chatting",
		},
		{
			Name:        "leave",
			Description: "Just got paged",
		},
		{
			Name:        "stream",
			Description: "Stream a youtube url to the voice channel",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "url",
					Description: "youtube url",
					Required:    true,
				},
			},
		},
		{
			Name:        "cease",
			Description: "Cease playback of the current song",
		},
		{
			Name:        "recess",
			Description: "Toggle a brief recess in playback",
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
					rss.ItemToEmbed(item, images, ""),
				},
			})
		},
		"join": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Joining voice channel...",
					// Flags:   1 << 6, // Ephemeral reply
				},
			})
			if err != nil {
				s.FollowupMessageCreate(s.State.User.ID, i.Interaction, true, &discordgo.WebhookParams{
					Content: "Something went wrong",
				})
				return
			}
			if i.Member == nil {
				s.InteractionResponseEdit(s.State.User.ID, i.Interaction, &discordgo.WebhookEdit{
					Content: "This command is only valid in guilds",
				})
				return
			}
			vs, err := s.State.VoiceState(i.GuildID, i.Member.User.ID)
			if err != nil {
				log.Println(err)
				s.InteractionResponseEdit(s.State.User.ID, i.Interaction, &discordgo.WebhookEdit{
					Content: "Could not grab voice state for user",
				})
				return
			}
			err = voice.JoinVoice(s, i.GuildID, vs.ChannelID)
			if err != nil {
				log.Println(err)
				s.InteractionResponseEdit(s.State.User.ID, i.Interaction, &discordgo.WebhookEdit{
					Content: "Could not join voice channel",
				})
				return
			}
			s.InteractionResponseEdit(s.State.User.ID, i.Interaction, &discordgo.WebhookEdit{
				Content: "Joined channel",
			})
		},
		"leave": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
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
			err = voice.LeaveVoice(i.GuildID)
			if err != nil {
				log.Println(err)
				s.InteractionResponseEdit(s.State.User.ID, i.Interaction, &discordgo.WebhookEdit{
					Content: "Failed to leave voice channel",
				})
				return
			}
			s.InteractionResponseEdit(s.State.User.ID, i.Interaction, &discordgo.WebhookEdit{
				Content: "Getting paged",
			})
		},
		"stream": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
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
			info, err := voice.UrlToEmbed(i.ApplicationCommandData().Options[0].StringValue())
			if err != nil {
				s.InteractionResponseEdit(s.State.User.ID, i.Interaction, &discordgo.WebhookEdit{
					Content: "Failed to retrieve video info",
				})
			} else {
				s.InteractionResponseEdit(s.State.User.ID, i.Interaction, &discordgo.WebhookEdit{
					Embeds: []*discordgo.MessageEmbed{info},
				})
			}
			go func(s *discordgo.Session, i *discordgo.InteractionCreate, url, gID string, info *discordgo.MessageEmbed) {
				s.UpdateListeningStatus(info.Title)
				defer s.UpdateListeningStatus("")
				err := voice.StreamUrl(url, gID)
				if err != nil {
					info.Author = &discordgo.MessageEmbedAuthor{
						Name: fmt.Sprintf("Error occured during playback: \n%v", err),
					}
					s.InteractionResponseEdit(s.State.User.ID, i.Interaction, &discordgo.WebhookEdit{
						Embeds: []*discordgo.MessageEmbed{info},
					})
					return
				}
				info.Author = &discordgo.MessageEmbedAuthor{
					Name: "Finished playing",
				}
				s.InteractionResponseEdit(s.State.User.ID, i.Interaction, &discordgo.WebhookEdit{
					Embeds: []*discordgo.MessageEmbed{info},
				})
			}(s, i, i.ApplicationCommandData().Options[0].StringValue(), i.GuildID, info)
		},
		"cease": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
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
			err = voice.Skip(i.GuildID)
			if err != nil {
				s.InteractionResponseEdit(s.State.User.ID, i.Interaction, &discordgo.WebhookEdit{
					Content: "An error occurred while attempting to skip",
				})
				return
			}
			s.InteractionResponseEdit(s.State.User.ID, i.Interaction, &discordgo.WebhookEdit{
				Content: "Successfully skipped",
			})
		},
		"recess": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
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
			err = voice.Pause(i.GuildID)
			if err != nil {
				s.InteractionResponseEdit(s.State.User.ID, i.Interaction, &discordgo.WebhookEdit{
					Content: "An error occured while attempting to pause",
				})
				return
			}
			s.InteractionResponseEdit(s.State.User.ID, i.Interaction, &discordgo.WebhookEdit{
				Content: "Successfully toggled pause",
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
		log.Println("bot is running...")
	})
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
	ug, err := session.UserGuilds(100, "", "")
	if err != nil {
		log.Panicf("could not retrieve user guilds: %v", err)
	}
	for _, v := range commands {
		for _, g := range ug {
			_, err := session.ApplicationCommandCreate(session.State.User.ID, g.ID, v)
			if err != nil {
				log.Panicf("could not create '%v' command: %v", v.Name, err)
			}
		}

	}
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop
	log.Println("shutting down...")
	if *RemoveCommands {
		for _, g := range ug {
			cmds, _ := session.ApplicationCommands(session.State.User.ID, g.ID)
			for _, v := range cmds {
				err := session.ApplicationCommandDelete(session.State.User.ID, g.ID, v.ID)
				if err != nil {
					log.Panicf("could not delete '%v' command (id %v): %v", v.Name, v.ID, err)
				}
			}
		}
	}
	var signal interface{}
	done <- signal
	<-ret
}
