package homework

import (
	"fmt"
	"io/ioutil"
	"math/rand"
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

var getProblemByMessageContent map[string]*problem

func init() {
	getProblemByMessageContent = map[string]*problem{}
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
		return "", fmt.Errorf("could not read directory %v: %v", PROBLEMS_DIR, err)
	}
	randomFile := files[rand.Intn(len(files))].Name()
	content, err := ioutil.ReadFile(dir + "/" + randomFile)
	if err != nil {
		return "", fmt.Errorf("could not read file %v: %v", randomFile, err)
	}
	return string(content), nil
}

// Create the appropriate discord message content from a markdown problem
func markdownProblemToMessageContent(p *problem) string {
	return fmt.Sprintf("**%s**\n\n%s", strings.TrimSpace(p.Title), p.Body)
}

// Fetch extract the title from markdown problem message content
func messageContentToProblemContent(content string) string {
	return strings.Join(strings.SplitAfter(content, "\n")[2:], "")
}

// Construct a discord SelectMenu from the current cached problems
func getSelectMenuOptionsFromCachedProblems() []discordgo.SelectMenuOption {
	result := []discordgo.SelectMenuOption{}
	for _, p := range getProblemByMessageContent {
		result = append(result, discordgo.SelectMenuOption{
			Label:       p.Title,
			Value:       markdownProblemToMessageContent(p),
			Default:     false,
			Description: "Show the solution for this problem",
		})
	}
	return result
}
