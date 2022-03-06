// package code provides functions and commands to pretty print code
package code

import (
	"log"
	"os"
	"strconv"

	"github.com/bwmarrin/discordgo"
	"github.com/harvlin/godot/module"
)

var commands []*discordgo.ApplicationCommand
var commandHandlers map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate)

func init() {
	commands = []*discordgo.ApplicationCommand{
		{
			Name: "echo",
			Type: discordgo.MessageApplicationCommand,
		},
		{
			Name: "code-block",
			Type: discordgo.MessageApplicationCommand,
		},
	}
	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"echo": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: i.ApplicationCommandData().Resolved.Messages[i.ApplicationCommandData().TargetID].Content,
				},
			})
			if err != nil {
				log.Printf("echo: %v", err)
			}
		},
		"code-block": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "",
				},
			})
			if err != nil {
				log.Printf("code-block: failed to respond to interaction: %v", err)
				return
			}
			content := i.ApplicationCommandData().Resolved.Messages[i.ApplicationCommandData().TargetID].Content
			codeBlocks := findCodeBlocks(content)
			if len(codeBlocks) == 0 {
				_, err = s.InteractionResponseEdit(s.State.User.ID, i.Interaction, &discordgo.WebhookEdit{
					Content: "This message has no code blocks :page_facing_up:",
				})
				return
			}
			files := []*os.File{}
			for i, block := range codeBlocks {
				outFile, err := makeCodeBlockImage(block, "-"+strconv.Itoa(i))
				if err != nil {
					log.Printf("code-block: failed to write to output file: %v", err)
					return
				}
				f, err := os.Open(outFile)
				defer f.Close()
				if err != nil {
					log.Printf("code-block: failed to open image: %v", err)
					return
				}
				files = append(files, f)
			}
			discordFiles := []*discordgo.File{}
			for _, f := range files {
				discordFiles = append(discordFiles, &discordgo.File{
					Name:        "code-block.png",
					ContentType: "multipart/form-data",
					Reader:      f,
				})
			}
			_, err = s.InteractionResponseEdit(s.State.User.ID, i.Interaction, &discordgo.WebhookEdit{
				Files: discordFiles,
			})
			if err != nil {
				log.Printf("code-block: %v", err)
			}
		},
	}
}

// GetModule returns the command Module for voice features
func GetModule() module.Module {
	return module.CreateModule(commands, commandHandlers, map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){})
}
