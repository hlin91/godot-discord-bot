package rss

import (
	"fmt"
	"log"
	"net/http"

	"github.com/mmcdole/gofeed"
	"golang.org/x/net/html"
)

const (
	MAX_ITEMS = 5
)

type Feed struct {
	Url       string
	Class     string
	NumImages int
}

var feeds []Feed                 // List of feeds to parse
var seen map[Feed][]*gofeed.Item // Remember the items we have already seen
var parser *gofeed.Parser

func init() {
	feeds = []Feed{}
	seen = map[Feed][]*gofeed.Item{}
	parser = gofeed.NewParser()
}

// ClearHistory clears the recently seen lists
func ClearHistory() {
	for key, _ := range seen {
		seen[key] = []*gofeed.Item{}
	}
}

// AddFeed adds a url to the list of feeds to parse
func AddFeed(url string, class string, n int) {
	feeds = append(feeds, Feed{
		Url:       url,
		Class:     class,
		NumImages: n,
	})
}

// GetLatest gets the latest items, up to MAX_ITEMS, that have not been seen during its last call
func GetLatest() map[Feed][]*gofeed.Item {
	result := map[Feed][]*gofeed.Item{}
	for _, f := range feeds {
		feed, err := parser.ParseURL(f.Url)
		if err != nil {
			log.Print(err)
			continue
		}
		items := feed.Items[0:MAX_ITEMS]
		for _, i := range items {
			if !itemInList(seen[f], i) {
				result[f] = append(result[f], i)
			}
		}
	}
	// Update our seen items
	for feed, items := range result {
		// Push the new items to the back of the seen list and rotate left accordingly
		seen[feed] = append(items[len(result[feed]):], result[feed]...)
	}
	return result
}

// GetImages returns the first n images from the given url page. If a root class is provided, all
// nodes with that class will be searched for images
func GetImages(url, class string, n int) ([]string, error) {
	result := []string{}
	// Load the page and parse the html
	resp, err := http.Get(url)
	if err != nil {
		return []string{}, fmt.Errorf("getting %s: %v", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return []string{}, fmt.Errorf("getting %s: %v", url, resp.StatusCode)
	}
	doc, err := html.Parse(resp.Body)
	if err != nil {
		return []string{}, fmt.Errorf("parsing %s as HTML: %v", url, err)
	}
	if class == "" {
		result = getImagesHelp(doc, "a", result, n)
		result = getImagesHelp(doc, "img", result, n)
		return result, nil
	}
	nodes := []*html.Node{}
	nodes = getNodesByClass(doc, class, nodes)
	for _, node := range nodes {
		result = getImagesHelp(node, "a", result, n)
	}
	for _, node := range nodes {
		result = getImagesHelp(node, "img", result, n)
	}
	return result, nil
}
