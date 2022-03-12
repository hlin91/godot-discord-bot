package homework

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/andybalholm/cascadia"
	"github.com/bwmarrin/discordgo"
	"golang.org/x/net/html"
)

const (
	// Leetcode related constants
	LEETCODE_PAGE_ROOT          = `https://leetcode.com`
	LEETCODE_PROBLEM_SET        = `https://leetcode.com/problemset/all/`
	LEETCODE_NAV_BUTTONS        = `nav[role="navigation"] button`
	LEETCODE_PROBLEM_ANCHORS    = `div[role="rowgroup"] div[role="cell"] a:not([aria-label="solution"])`
	LEETCODE_PROBLEM_DIFFICULTY = `div[diff]`
	USER_AGENT                  = `Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:97.0) Gecko/20100101 Firefox/97.0`
	LEETCODE_ICON_URL           = `https://avatars0.githubusercontent.com/u/48126122?s=280&v=4`
	// Rosetta Code related constants
	ROSETTA_PAGE_ROOT       = `https://rosettacode.org`
	ROSETTA_PROBLEM_SET     = `https://rosettacode.org/wiki/Category:Programming_Tasks`
	ROSETTA_PROBLEM_ANCHORS = `div.mw-category-group ul li a`
	POP_QUIZ_ICON_URL       = `https://github.com/cat-milk/Anime-Girls-Holding-Programming-Books/blob/master/C++/cirno_teaches_c++.jpg?raw=true`
)

type myJar struct {
	jar map[string][]*http.Cookie
}

func (p *myJar) SetCookies(u *url.URL, cookies []*http.Cookie) {
	log.Printf("The URL is : %s\n", u.String())
	log.Printf("The cookie being set is : %s\n", cookies)
	p.jar[u.Host] = cookies
}

func (p *myJar) Cookies(u *url.URL) []*http.Cookie {
	log.Printf("The URL is : %s\n", u.String())
	log.Printf("Cookie being returned is : %s\n", p.jar[u.Host])
	return p.jar[u.Host]
}

func httpClientWithCookieJar() *http.Client {
	// Set up the cookie jar so requests can be authenticated
	jar := &myJar{}
	jar.jar = map[string][]*http.Cookie{}
	client := &http.Client{}
	client.Jar = jar
	return client
}

// Return the upper limit of pages of problems on leetcode
func leetcodeGetNumPages() (int, error) {
	// Hard code this in for now until we know how to dynamically determine this
	// Probably not possible since the site is 100% javascript
	return 44, nil

	// client := httpClientWithCookieJar()
	// req, err := http.NewRequest("GET", LEETCODE_PAGE_ROOT, nil)
	// if err != nil {
	// 	return 0, fmt.Errorf("making request: %v", err)
	// }
	// req.Header.Set("User-Agent", USER_AGENT)

	// resp, err := client.Do(req)
	// if err != nil {
	// 	return 0, fmt.Errorf("getting %s: %v", LEETCODE_PAGE_ROOT, err)
	// }
	// defer resp.Body.Close()
	// if resp.StatusCode != http.StatusOK {
	// 	return 0, fmt.Errorf("getting %s: %v", LEETCODE_PAGE_ROOT, err)
	// }
	// doc, err := html.Parse(resp.Body)
	// if err != nil {
	// 	return 0, fmt.Errorf("parsing %s as HTML: %v", LEETCODE_PAGE_ROOT, err)
	// }
	// buttons := cascadia.QueryAll(doc, cascadia.MustCompile(LEETCODE_NAV_BUTTONS))
	// if len(buttons) <= 1 {
	// 	return len(buttons), nil
	// }
	// return strconv.Atoi(buttons[len(buttons)-2].FirstChild.Data)
}

// Get the list of problem urls for a given page
// By default, leetcode returns 50 problems per page
func leetcodeGetProblemListForPage(page int) ([]string, error) {
	if page <= 0 {
		return []string{}, fmt.Errorf("page must be greater than 0")
	}
	result := []string{}
	url := LEETCODE_PROBLEM_SET + fmt.Sprintf("?page=%d", page)
	// Construct and send the http GET request
	client := httpClientWithCookieJar()
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return []string{}, fmt.Errorf("making request: %v", err)
	}
	req.Header.Set("User-Agent", USER_AGENT)
	resp, err := client.Do(req)
	if err != nil {
		return []string{}, fmt.Errorf("getting %s: %v", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return []string{}, fmt.Errorf("getting %s: %v", url, err)
	}
	doc, err := html.Parse(resp.Body)
	if err != nil {
		return []string{}, fmt.Errorf("parsing %s as HTML: %v", url, err)
	}
	// Parse for the anchor elements containing the problem links
	anchors := cascadia.QueryAll(doc, cascadia.MustCompile(LEETCODE_PROBLEM_ANCHORS))
	for _, a := range anchors {
		for _, attr := range a.Attr {
			if attr.Key == "href" {
				result = append(result, attr.Val)
			}
		}
	}
	return result, nil
}

// Return the url to a random question on leetcode
func leetcodeGetRandomQuestion() (string, error) {
	rand.Seed(time.Now().Unix())
	numPages, err := leetcodeGetNumPages()
	if err != nil {
		return "", fmt.Errorf("failed to retrieve number of pages: %v", err)
	}
	page := rand.Intn(numPages) + 1
	problems, err := leetcodeGetProblemListForPage(page)
	if err != nil {
		return "", fmt.Errorf("failed to retrieve problem list for page %v: %v", page, err)
	}
	randomProblem := problems[rand.Intn(len(problems))]
	return LEETCODE_PAGE_ROOT + randomProblem, nil
}

// Construct a discord message embed from a leetcode question link
func leetcodeQuestionToEmbed(url string) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		URL:         url,
		Type:        discordgo.EmbedTypeArticle,
		Title:       ":books: Homework Assignment :books:",
		Description: "Due tomorrow!",
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: LEETCODE_ICON_URL,
		},
		Color: 0xFF9900,
	}
}

// Return the list of coding problems for rosetta
func rosettaGetProblemList() ([]string, error) {
	url := ROSETTA_PROBLEM_SET
	client := httpClientWithCookieJar()
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return []string{}, fmt.Errorf("making request: %v", err)
	}
	req.Header.Set("User-Agent", USER_AGENT)
	resp, err := client.Do(req)
	if err != nil {
		return []string{}, fmt.Errorf("getting %s: %v", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return []string{}, fmt.Errorf("getting %s: %v", url, err)
	}
	doc, err := html.Parse(resp.Body)
	if err != nil {
		return []string{}, fmt.Errorf("parsing %s as HTML: %v", url, err)
	}
	anchors := cascadia.QueryAll(doc, cascadia.MustCompile(ROSETTA_PROBLEM_ANCHORS))
	result := []string{}
	for _, a := range anchors {
		for _, attr := range a.Attr {
			if attr.Key == "href" {
				result = append(result, attr.Val)
			}
		}
	}
	return result, nil
}

// Return the link to a random question on rosetta
func rosettaGetRandomQuestion() (string, error) {
	rand.Seed(time.Now().Unix())
	problems, err := rosettaGetProblemList()
	if err != nil {
		return "", fmt.Errorf("failed to retrieve problem list: %v", err)
	}
	randomProblem := problems[rand.Intn(len(problems))]
	return ROSETTA_PAGE_ROOT + randomProblem, nil
}

// Calculate the edit distance between 2 strings
func editDistance(s1, s2 string) int {
	if len(s1) == 0 {
		return len(s2)
	}
	if len(s2) == 0 {
		return len(s1)
	}
	s1 = strings.ToLower(s1)
	s2 = strings.ToLower(s2)
	dp := [][]int{}
	for i := 0; i <= len(s1); i++ {
		dp = append(dp, make([]int, len(s2)+1))
	}
	for i := range dp {
		for j := range dp[i] {
			if i == 0 {
				dp[i][j] = j
				continue
			}
			if j == 0 {
				dp[i][j] = i
				continue
			}
			cost := 1
			if s1[i-1] == s2[j-1] {
				cost = 0
			}
			dp[i][j] = int(math.Min(math.Min(float64(dp[i-1][j]+1), float64(dp[i][j-1]+1)), float64(dp[i-1][j-1]+cost)))
		}
	}
	return dp[len(s1)-1][len(s2)-1]
}
