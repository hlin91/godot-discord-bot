package voice

import (
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/harvlin/godot/module"
)

var commands []*discordgo.ApplicationCommand
var commandHandlers map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate)

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
			info, err := UrlToEmbed(i.ApplicationCommandData().Options[0].StringValue())
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
				var signal interface{}
				StatusLock <- signal
				s.UpdateListeningStatus(info.Title)
				defer func(s *discordgo.Session) {
					s.UpdateListeningStatus("")
					<-StatusLock
				}(s)
				err := StreamUrl(url, gID)
				if err != nil {
					info.Author = &discordgo.MessageEmbedAuthor{
						Name: fmt.Sprintf("Error occurred during playback: \n%v", err),
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
				Content: "Playback has ceased",
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
	}
}

// GetModule returns the command Module for voice features
func GetModule() module.Module {
	return module.CreateModule(commands, commandHandlers, map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){})
}
