package homework

import (
	"fmt"
	"log"
	"sort"

	"github.com/bwmarrin/discordgo"
	"github.com/harvlin/godot/module"
)

var commands []*discordgo.ApplicationCommand
var commandHandlers map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate)
var componentHandlers map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate)
var modalHandlers map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate)

func init() {
	commands = []*discordgo.ApplicationCommand{
		{
			Name: "assign-homework",
			Type: discordgo.UserApplicationCommand,
		},
		{
			Name: "pop-quiz",
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
				return
			}
			questionUrl, err := leetcodeGetRandomQuestion()
			if err != nil {
				if err != nil {
					log.Printf("assign-homework: failed to retrieve question: %v", err)
					return
				}
			}
			_, err = s.InteractionResponseEdit(s.State.User.ID, i.Interaction, &discordgo.WebhookEdit{
				Content: fmt.Sprintf("<@%v>", i.ApplicationCommandData().TargetID),
				Embeds:  []*discordgo.MessageEmbed{leetcodeQuestionToEmbed(questionUrl)},
			})
			if err != nil {
				log.Printf("assign-homework: failed to edit interaction: %v", err)
			}
		},
		"pop-quiz": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "",
				},
			})
			if err != nil {
				log.Printf("pop-quiz: failed to respond to interaction: %v", err)
				return
			}
			problemContent, err := fetchRandomMarkdownProblemContent(PROBLEMS_DIR)
			if err != nil {
				if err != nil {
					log.Printf("pop-quiz: failed to fetch problem: %v", err)
					return
				}
			}
			problem := createProblem(problemContent)
			getProblemByInteractionId[i.Interaction.ID] = problem
			_, err = s.InteractionResponseEdit(s.State.User.ID, i.Interaction, &discordgo.WebhookEdit{
				Content: fmt.Sprintf("<@%v>\n***Pop quiz!***\n", i.ApplicationCommandData().TargetID) + markdownProblemToMessageContent(problem),
				Components: []discordgo.MessageComponent{
					discordgo.ActionsRow{
						Components: []discordgo.MessageComponent{
							discordgo.Button{
								Label:    "Show me!",
								Style:    discordgo.DangerButton,
								Disabled: false,
								CustomID: "show_solution",
							},
						},
					},
				},
			})
			if err != nil {
				log.Printf("pop-quiz: failed to edit interaction: %v", err)
			}
		},
	}
	componentHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"search_solutions": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseModal,
				Data: &discordgo.InteractionResponseData{
					Content:  "Which solution would you like to see? :mag:",
					CustomID: "search_solution",
					Title:    "Solution searcher",
					Components: []discordgo.MessageComponent{
						discordgo.ActionsRow{
							Components: []discordgo.MessageComponent{
								discordgo.TextInput{
									CustomID:    "search_solution",
									Placeholder: "Enter solution title here...",
									Label:       "Search for a solution by title",
									Style:       discordgo.TextInputShort,
									Required:    true,
									MinLength:   1,
									MaxLength:   500,
								},
							},
						},
					},
				},
			})
			if err != nil {
				log.Printf("search_solutions: failed to respond to interaction: %v", err)
			}
		},
		"show_solution": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			problem, ok := getProblemByInteractionId[i.Message.Interaction.ID]
			if !ok {
				err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "That solution is not available anymore :floppy_disk: Would you like to look it up? :mag:",
						Components: []discordgo.MessageComponent{
							discordgo.ActionsRow{
								Components: []discordgo.MessageComponent{
									discordgo.Button{
										Label:    "Yes",
										Style:    discordgo.SuccessButton,
										Disabled: false,
										CustomID: "search_solutions",
									},
								},
							},
						},
					},
				})
				if err != nil {
					log.Printf("show_solution: failed to respond: %v", err)
				}
				return
			}
			showSolutionHelp(problem, s, i)
			// delete(getProblemByInteractionId, i.Message.Interaction.ID)
		},
	}
	modalHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"search_solution": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Looking up solution... :mag:",
				},
			})
			if err != nil {
				log.Printf("hard_lookup_solution: failed to respond to interaction: %v", err)
			}
			problemBank, err := getAllProblemsSingleton()
			if err != nil {
				log.Printf("hard_lookup_solution: failed to retrieve singleton: %v", err)
			}
			input := i.ModalSubmitData().Components[0].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value
			keys := []string{}
			for k := range problemBank {
				keys = append(keys, k)
			}
			sort.Slice(keys, func(i, j int) bool {
				return editDistance(keys[i], input) < editDistance(keys[j], input)
			})
			response, err := solutionToResponseData(problemBank[keys[0]])
			if err != nil {
				log.Printf("hard_lookup_solution: failed to generate response: %v", err)
			}
			_, err = s.InteractionResponseEdit(s.State.User.ID, i.Interaction, &discordgo.WebhookEdit{
				Content: response.Content,
				Files:   response.Files,
			})
			if err != nil {
				log.Printf("hard_lookup_solution: failed to edit response: %v", err)
			}
		},
	}
}

func GetModule() module.Module {
	return module.CreateModule(commands, commandHandlers, componentHandlers, modalHandlers)
}
