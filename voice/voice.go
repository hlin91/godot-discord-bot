package voice

import (
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/harvlin/godot/module"
)

var commands []*discordgo.ApplicationCommand
var commandHandlers map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate)
var componentHandlers map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate)

var StatusLock = make(chan interface{}, 1)

func init() {
	commands = []*discordgo.ApplicationCommand{
		{
			Name:        "join",
			Description: "Just chatting",
		},
		{
			Name:        "leave",
			Description: "Page me",
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
		{
			Name:        "recent",
			Description: "Stream a recently played song",
		},
	}
	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"join": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Joining voice channel...",
					// Flags:   1 << 6, // Ephemeral reply
				},
			})
			if err != nil {
				log.Printf("join: failed to respond to interaction: %v", err)
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
			err = JoinVoice(s, i.GuildID, vs.ChannelID)
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
				log.Printf("leave: failed to respond to interaction: %v", err)
				return
			}
			err = LeaveVoice(i.GuildID)
			s.UpdateListeningStatus("")
			if err != nil {
				log.Println(err)
				s.InteractionResponseEdit(s.State.User.ID, i.Interaction, &discordgo.WebhookEdit{
					Content: "Failed to leave voice channel",
				})
				return
			}
			s.InteractionResponseEdit(s.State.User.ID, i.Interaction, &discordgo.WebhookEdit{
				Content: ":alarm_clock: Getting paged! :alarm_clock:",
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
				log.Printf("stream: failed to respond to interaction: %v", err)
				return
			}
			url := i.ApplicationCommandData().Options[0].StringValue()
			info, err := UrlToEmbed(url)
			if err != nil {
				s.InteractionResponseEdit(s.State.User.ID, i.Interaction, &discordgo.WebhookEdit{
					Content: "Failed to retrieve video info",
				})
			} else {
				s.InteractionResponseEdit(s.State.User.ID, i.Interaction, &discordgo.WebhookEdit{
					Embeds: []*discordgo.MessageEmbed{info},
				})
			}
			recentlyPlayed[info.Title] = url
			go streamUrlCoroutine(s, i, url, i.GuildID, info)
		},
		"cease": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "",
				},
			})
			if err != nil {
				log.Printf("cease: failed to respond to interaction: %v", err)
				return
			}
			err = Skip(i.GuildID)
			if err != nil {
				s.InteractionResponseEdit(s.State.User.ID, i.Interaction, &discordgo.WebhookEdit{
					Content: "An error occurred while attempting to skip",
				})
				return
			}
			s.InteractionResponseEdit(s.State.User.ID, i.Interaction, &discordgo.WebhookEdit{
				Content: "Playback has ceased for the current song",
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
				log.Printf("recess: failed to respond to interaction: %v", err)
				return
			}
			err = Pause(i.GuildID)
			if err != nil {
				s.InteractionResponseEdit(s.State.User.ID, i.Interaction, &discordgo.WebhookEdit{
					Content: "An error occurred while attempting to pause",
				})
				return
			}
			s.InteractionResponseEdit(s.State.User.ID, i.Interaction, &discordgo.WebhookEdit{
				Content: "Successfully toggled pause",
			})
		},
		"recent": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			if len(recentlyPlayed) == 0 {
				err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "No songs have been played recently :musical_score:",
					},
				})
				if err != nil {
					log.Printf("list_solutions: failed to respond to interaction: %v", err)
				}
				return
			}
			err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "",
				},
			})
			if err != nil {
				log.Printf("recent: failed to respond to interaction: %v", err)
				return
			}
			selectMenuOptions := getSelectMenuOptionsFromRecentlyPlayed()
			if len(selectMenuOptions) == 0 {
				log.Printf("recent: warning: selectMenuOptions is empty")
			}
			_, err = s.InteractionResponseEdit(s.State.User.ID, i.Interaction, &discordgo.WebhookEdit{
				Content: "Select a track :dvd::musical_note:",
				Components: []discordgo.MessageComponent{
					discordgo.ActionsRow{
						Components: []discordgo.MessageComponent{
							discordgo.SelectMenu{
								CustomID:    "play_selection",
								Placeholder: "Song title",
								Options:     selectMenuOptions,
							},
						},
					},
				},
			})
			if err != nil {
				log.Printf("list_solutions: failed to edit interaction: %v", err)
			}
		},
	}
	componentHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"play_selection": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "",
				},
			})
			if err != nil {
				log.Printf("play_selection: failed to respond to interaction: %v", err)
				return
			}
			url := i.MessageComponentData().Values[0]
			info, err := UrlToEmbed(url)
			if err != nil {
				s.InteractionResponseEdit(s.State.User.ID, i.Interaction, &discordgo.WebhookEdit{
					Content: "Failed to retrieve video info",
				})
			} else {
				s.InteractionResponseEdit(s.State.User.ID, i.Interaction, &discordgo.WebhookEdit{
					Embeds: []*discordgo.MessageEmbed{info},
				})
			}
			go streamUrlCoroutine(s, i, url, i.GuildID, info)
		},
	}
}

// GetModule returns the command Module for voice features
func GetModule() module.Module {
	return module.CreateModule(commands, commandHandlers, componentHandlers)
}
