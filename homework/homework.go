package homework

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"log"
	"os"

	"github.com/bwmarrin/discordgo"
	"github.com/harvlin/godot/module"
)

var commands []*discordgo.ApplicationCommand
var commandHandlers map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate)
var componentHandlers map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate)

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
			getProblemByTitle[problem.Title] = problem
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
		"list_solutions": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			if len(getProblemByTitle) == 0 {
				err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "Sorry, I have no more solutions to show :pensive:",
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
				log.Printf("list_solutions: failed to respond to interaction: %v", err)
				return
			}
			selectMenuOptions := getSelectMenuOptionsFromCachedProblems()
			if len(selectMenuOptions) == 0 {
				log.Printf("list_solutions: warning: selectMenuOptions is empty")
			}
			_, err = s.InteractionResponseEdit(s.State.User.ID, i.Interaction, &discordgo.WebhookEdit{
				Content: "Which solution do you want to see?",
				Components: []discordgo.MessageComponent{
					discordgo.ActionsRow{
						Components: []discordgo.MessageComponent{
							discordgo.SelectMenu{
								CustomID:    "show_solution",
								Placeholder: "Choose a solution",
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
		"show_solution": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			title := markdownProblemMessageContentToTitle(i.Message.Content)
			problem, ok := getProblemByTitle[title]
			if !ok {
				err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "That solution is not available anymore :floppy_disk:",
					},
				})
				if err != nil {
					log.Printf("show_solution: failed to respond: %v", err)
				}
				return
			}
			solution := problem.Solution
			if len(solution) > 2000 {
				// Solution is over 2000 chars so we have to upload it as a file
				err := ioutil.WriteFile(TMP_SOLUTION_FILE, []byte(solution), fs.ModePerm)
				if err != nil {
					log.Printf("show_solution: failed to write file %v: %v", TMP_SOLUTION_FILE, err)
					return
				}
				file, err := os.Open(TMP_SOLUTION_FILE)
				if err != nil {
					log.Printf("show_solution: failed to open file %v: %v", TMP_SOLUTION_FILE, err)
					return
				}
				err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "This solution is too large. Uploading as a file :file_folder:",
						Files: []*discordgo.File{
							{
								Name:        "solution.md",
								ContentType: "multipart/form-data",
								Reader:      file,
							},
						},
					},
				})
				if err != nil {
					log.Printf("show_solution: failed to respond to interaction: %v", err)
				}
				delete(getProblemByTitle, title)
				return
			}
			err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: solution,
				},
			})
			if err != nil {
				log.Printf("show_solution: failed to respond to interaction: %v", err)
			}
			delete(getProblemByTitle, title)
		},
	}
}

func GetModule() module.Module {
	return module.CreateModule(commands, commandHandlers, componentHandlers)
}
