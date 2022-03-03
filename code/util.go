package code

import (
	"context"
	"os"
	"regexp"
	"strings"

	"golang.design/x/code2img"
)

const (
	TEMP_IMG_FILE                = "/tmp/code_block_img.png"
	CODE_BLOCK_START_END_PATTERN = "^```[[:alpha:]]*$"
)

// Trim the start and end indicators of a code block if present
func trimCodeBlock(s string) string {
	lines := strings.Split(s, "\n")
	result := []string{}
	for _, l := range lines {
		isIndicator, _ := regexp.MatchString(CODE_BLOCK_START_END_PATTERN, l)
		if isIndicator {
			continue
		}
		result = append(result, l)
	}
	return strings.Join(result, "\n")
}

// Create the code block image and save it as TEMP_IMG_FILE
func makeCodeBlock(content string) error {
	content = trimCodeBlock(content)
	img, err := code2img.Render(context.TODO(), code2img.LangAuto, content)
	if err != nil {
		return err
	}
	err = os.WriteFile(TEMP_IMG_FILE, img, os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}
