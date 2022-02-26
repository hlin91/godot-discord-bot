package homework

import (
	"fmt"
	"testing"
)

func TestGetNumPagesLeetcode(t *testing.T) {
	numPages, err := leetcodeGetNumPages()
	if err != nil {
		t.Errorf("%v", err)
	}
	if numPages == 0 {
		t.Errorf("Got numPages = %v, want >= 1", numPages)
	}
}

func TestGetProblemListForPage(t *testing.T) {
	result, err := leetcodeGetProblemListForPage(1)
	if err != nil {
		t.Errorf("error fetching problem list: %v", err)
	}
	if len(result) == 0 {
		t.Errorf("got 0 problems, want >= 1")
	}
	fmt.Printf("got %v problem(s): %v", len(result), result)
}

func TestGetRandomQuestion(t *testing.T) {
	link, err := leetcodeGetRandomQuestion()
	if err != nil {
		t.Errorf("%v", err)
	}
	if len(link) == 0 {
		t.Errorf("failed to retrieve a question")
	}
	fmt.Printf("got question link: %v\n", link)
}

func TestRosettaGetProblemList(t *testing.T) {
	result, err := rosettaGetProblemList()
	if err != nil {
		t.Errorf("error fetching problem list: %v", err)
	}
	if len(result) == 0 {
		t.Errorf("got 0 problems, want >= 1")
	}
	fmt.Printf("got %v problem(s): %v", len(result), result)
}

func TestRosettaGetRandomQuestion(t *testing.T) {
	result, err := rosettaGetRandomQuestion()
	if err != nil {
		t.Errorf("error fetching problem list: %v", err)
	}
	if len(result) == 0 {
		t.Errorf("got 0 problems, want >= 1")
	}
	fmt.Printf("got %v problem(s): %v", len(result), result)
}
