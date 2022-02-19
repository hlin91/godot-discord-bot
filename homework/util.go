package homework

import (
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"time"

	"github.com/andybalholm/cascadia"
	"golang.org/x/net/html"
)

const (
	LEETCODE_PAGE_ROOT          = `https://leetcode.com`
	LEETCODE_PROBLEM_SET        = `https://leetcode.com/problemset/all/`
	LEETCODE_NAV_BUTTONS        = `nav[role="navigation"] button`
	LEETCODE_PROBLEM_ANCHORS    = `div[role="rowgroup"] div[role="cell"] a:not([aria-label="solution"])`
	LEETCODE_PROBLEM_DIFFICULTY = `div[diff]`
	USER_AGENT                  = `Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:97.0) Gecko/20100101 Firefox/97.0`
)

type myJar struct {
	jar map[string][]*http.Cookie
}

func (p *myJar) SetCookies(u *url.URL, cookies []*http.Cookie) {
	fmt.Printf("The URL is : %s\n", u.String())
	fmt.Printf("The cookie being set is : %s\n", cookies)
	p.jar[u.Host] = cookies
}

func (p *myJar) Cookies(u *url.URL) []*http.Cookie {
	fmt.Printf("The URL is : %s\n", u.String())
	fmt.Printf("Cookie being returned is : %s\n", p.jar[u.Host])
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
