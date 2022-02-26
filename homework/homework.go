package homework

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/harvlin/godot/module"
)

var commands []*discordgo.ApplicationCommand
var commandHandlers map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate)

func init() {
	commands = []*discordgo.ApplicationCommand{
		{
			Name: "assign-homework",
			Type: discordgo.UserApplicationCommand,
		},
	}
	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"assign-homework": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
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
			questionUrl, err := leetcodeGetRandomQuestion()
			if err != nil {
				if err != nil {
					s.FollowupMessageCreate(s.State.User.ID, i.Interaction, true, &discordgo.WebhookParams{
						Content: "Failed to retrieve question :(",
					})
					return
				}
			}
			_, err = s.InteractionResponseEdit(s.State.User.ID, i.Interaction, &discordgo.WebhookEdit{
				Content: fmt.Sprintf("<@%v>", i.ApplicationCommandData().TargetID),
				Embeds:  []*discordgo.MessageEmbed{leetcodeQuestionToEmbed(questionUrl)},
			})
		},
	}
}

func GetModule() module.Module {
	return module.CreateModule(commands, commandHandlers)
}
