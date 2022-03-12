package homework

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

const (
	PROBLEMS_DIR              = `./homework/markdown_coding_problems`
	MARKDOWN_PROBLEM_TITLE    = `(?m)^#[[:blank:]]*([^#])+$`
	MARKDOWN_PROBLEM_BODY     = `(?si)##[[:blank:]]*(The )?(problem)?(task)?(.*)##[[:blank:]]*(The )?solution` // The last line matched is not needed
	MARKDOWN_PROBLEM_SOLUTION = `(?si)##[[:blank:]]*(The )?solution(.*)`
	TMP_SOLUTION_FILE         = `/tmp/solution.md`
)

type problem struct {
	Title    string
	Body     string
	Solution string
}

var getProblemByInteractionId map[string]*problem
var allProblems map[string]*problem

func init() {
	getProblemByInteractionId = map[string]*problem{}
	allProblems = map[string]*problem{}
}

// Construct a problem struct from markdown content fed as a string
func createProblem(s string) *problem {
	titleRegex := regexp.MustCompile(MARKDOWN_PROBLEM_TITLE)
	bodyRegex := regexp.MustCompile(MARKDOWN_PROBLEM_BODY)
	solutionRegex := regexp.MustCompile(MARKDOWN_PROBLEM_SOLUTION)

	title := titleRegex.FindString(s)
	body := bodyRegex.FindString(s)
	solution := solutionRegex.FindString(s)

	return &problem{
		Title:    title,
		Body:     body,
		Solution: solution,
	}
}

// Fetch a random markdown problem from the directory and return its content as a string
func fetchRandomMarkdownProblemContent(dir string) (string, error) {
	rand.Seed(time.Now().Unix())
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return "", fmt.Errorf("could not read directory %v: %v", dir, err)
	}
	randomFile := files[rand.Intn(len(files))].Name()
	content, err := ioutil.ReadFile(dir + "/" + randomFile)
	if err != nil {
		return "", fmt.Errorf("could not read file %v: %v", randomFile, err)
	}
	return string(content), nil
}

// Construct and return a map of all markdown problems in a directory mapped by problem title
func getAllProblemsInDir(dir string) (map[string]*problem, error) {
	result := map[string]*problem{}
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return map[string]*problem{}, fmt.Errorf("could not read directory %v: %v", dir, err)
	}
	for _, f := range files {
		content, err := ioutil.ReadFile(dir + "/" + f.Name())
		if err != nil {
			return map[string]*problem{}, fmt.Errorf("could not read file %v: %v", f, err)
		}
		p := createProblem(string(content))
		result[p.Title] = p
	}
	return result, nil
}

// Get the map of all locally stored problems as a singleton
func getAllProblemsSingleton() (map[string]*problem, error) {
	if len(allProblems) == 0 {
		problems, err := getAllProblemsInDir(PROBLEMS_DIR)
		if err != nil {
			return map[string]*problem{}, fmt.Errorf("getAllProblemsSingleton: failed to fetch problems: %v", err)
		}
		allProblems = problems
	}
	return allProblems, nil
}

// Create the appropriate discord message content from a markdown problem
func markdownProblemToMessageContent(p *problem) string {
	return fmt.Sprintf("**%s**\n\n%s", strings.TrimSpace(p.Title), p.Body)
}

// Construct a discord SelectMenu from the current cached problems
func getSelectMenuOptionsFromProblemMap(m map[string]*problem) []discordgo.SelectMenuOption {
	result := []discordgo.SelectMenuOption{}
	for key, p := range m {
		result = append(result, discordgo.SelectMenuOption{
			Label:       p.Title,
			Value:       key,
			Default:     false,
			Description: "Show the solution for this problem",
		})
	}
	return result
}

// Helper function for responding to a message with the problem solution
func showSolutionHelp(problem *problem, s *discordgo.Session, i *discordgo.InteractionCreate) {
	response, err := solutionToResponseData(problem)
	if err != nil {
		log.Printf("showSolutionHelp: failed to generate response data: %v", err)
	}
	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: response,
	})
	if err != nil {
		log.Printf("showSolutionHelp: failed to respond to interaction: %v", err)
	}
}

// Helper function for creating the appropriate InteractionResponseData from a problem solution
func solutionToResponseData(p *problem) (*discordgo.InteractionResponseData, error) {
	solution := p.Solution
	if len(solution) > 2000 {
		// Solution is over 2000 chars so we have to upload it as a file
		err := ioutil.WriteFile(TMP_SOLUTION_FILE, []byte(solution), fs.ModePerm)
		if err != nil {
			return nil, fmt.Errorf("solutionToResponseData: failed to write file %v: %v", TMP_SOLUTION_FILE, err)
		}
		file, err := os.Open(TMP_SOLUTION_FILE)
		if err != nil {
			return nil, fmt.Errorf("showSolutionHelp: failed to open file %v: %v", TMP_SOLUTION_FILE, err)
		}
		return &discordgo.InteractionResponseData{
			Content: "This solution is too large. Uploading as a file :file_folder:",
			Files: []*discordgo.File{
				{
					Name:        "solution.md",
					ContentType: "multipart/form-data",
					Reader:      file,
				},
			},
		}, nil
	}
	return &discordgo.InteractionResponseData{
		Content: solution,
	}, nil
}
