package code

import (
	"context"
	"log"
	"os"
	"regexp"
	"strings"

	"golang.design/x/code2img"
)

const (
	TEMP_IMG_FILE        = "/tmp/code_block_img.png"
	CODE_BLOCK_REGEX     = "(?s)```([[:alpha:]]*\\+*)?(.*?)```" // Group 1 is the language tag, group 2 is the body
	ONE_LINE_BLOCK_REGEX = "(?m)^```(.+)```$"                   // Group 1 is the body. There is no language tag
)

type codeBlock struct {
	Language string // The language indicator tag for the code block
	Body     string // The actual content of the code block
}

// Construct a codeblock object from a string
func createCodeBlock(s string) *codeBlock {
	isOneLine, err := regexp.MatchString(ONE_LINE_BLOCK_REGEX, s)
	if err != nil {
		log.Printf("createCodeBlock: failed to match string %v with pattern %v: %v", s, ONE_LINE_BLOCK_REGEX, err)
		return nil
	}
	if isOneLine {
		// One line code blocks do not have a language tag
		re := regexp.MustCompile(ONE_LINE_BLOCK_REGEX)
		groups := re.FindStringSubmatch(s)
		if groups == nil {
			log.Printf("createCodeBlock: failed to find submatches in string %v for pattern %v: %v", s, ONE_LINE_BLOCK_REGEX, err)
		}
		return &codeBlock{
			Language: "",
			Body:     strings.TrimSpace(groups[1]),
		}
	}
	re := regexp.MustCompile(CODE_BLOCK_REGEX)
	groups := re.FindStringSubmatch(s)
	if groups == nil {
		log.Printf("createCodeBlock: failed to find submatches in string %v for pattern %v: %v", s, ONE_LINE_BLOCK_REGEX, err)
	}
	return &codeBlock{
		Language: strings.TrimSpace(groups[1]),
		Body:     strings.TrimSpace(groups[2]),
	}
}

// Find the code blocks from the message content
func findCodeBlocks(content string) []*codeBlock {
	result := []*codeBlock{}
	re := regexp.MustCompile(CODE_BLOCK_REGEX)
	matches := re.FindAllString(content, -1)
	for _, s := range matches {
		result = append(result, createCodeBlock(s))
	}
	return result
}

// Create the code block image and save it as TEMP_IMG_FILE with a provided qualifier for the file name
// Returns the resulting file name
func makeCodeBlockImage(codeBlock *codeBlock, qualifier string) (string, error) {
	content := codeBlock.Body
	img, err := code2img.Render(context.TODO(), code2img.LangAuto, content)
	if err != nil {
		return "", err
	}
	filename := TEMP_IMG_FILE + qualifier
	err = os.WriteFile(filename, img, os.ModePerm)
	if err != nil {
		return "", err
	}
	return filename, nil
}
