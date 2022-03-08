package rss

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/harvlin/godot/module"
	"github.com/mmcdole/gofeed"
	"golang.org/x/net/html"
)

const (
	MAX_ITEMS  = 50
	USER_AGENT = `Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:97.0) Gecko/20100101 Firefox/97.0`
)

type Feed struct {
	Url                         string
	NumImages                   int
	ImageNodeFilterStrategy     *func(*html.Node) bool
	LogoImageNodeFilterStrategy *func(*html.Node) bool
	ImageLinkExtractionStrategy *func(*html.Node) string
	ImageLinkTransformStrategy  *func(string) string
	GetChannelIdStrategy        *func() string
}

var feeds []*Feed                 // List of feeds to parse
var seen map[*Feed][]*gofeed.Item // Remember the items we have already seen
var parser *gofeed.Parser

var commands []*discordgo.ApplicationCommand
var commandHandlers map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate)

func init() {
	feeds = []*Feed{}
	seen = map[*Feed][]*gofeed.Item{}
	parser = gofeed.NewParser()
	commands = []*discordgo.ApplicationCommand{
		{
			Name:        "test_rss",
			Description: "Test the RSS feed feature by posting an RSS post to the channel",
		},
	}
	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"test_rss": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "",
				},
			})
			if err != nil {
				log.Printf("test_rss: failed to respond to interaction: %v", err)
				return
			}
			// Do RSS stuff here
			ClearHistory()
			items := GetLatest()
			var images []string
			var logos []string
			embeds := []*discordgo.MessageEmbed{}
			for key, val := range items {
				images, _ = GetImages(val[0].Link, *key.ImageNodeFilterStrategy, key.NumImages, *key.ImageLinkExtractionStrategy, *key.ImageLinkTransformStrategy)
				logos, _ = GetImages(val[0].Link, *key.LogoImageNodeFilterStrategy, key.NumImages, *key.ImageLinkExtractionStrategy, *key.ImageLinkTransformStrategy)
				embeds = append(embeds, ItemToEmbed(val[0], images, logos))
			}
			_, err = s.InteractionResponseEdit(s.State.User.ID, i.Interaction, &discordgo.WebhookEdit{
				Embeds: embeds,
			})
			if err != nil {
				log.Printf("test_rss: %v", err)
			}
		},
	}
}

// GetModule returns the command Module for RSS features
func GetModule() module.Module {
	return module.CreateModule(commands, commandHandlers, map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){})
}

// ClearHistory clears the recently seen lists
func ClearHistory() {
	for key := range seen {
		seen[key] = []*gofeed.Item{}
	}
}

// AddFeed adds a url to the list of feeds to parse
func AddFeed(url string, imageNodeFilterStrategy func(*html.Node) bool, logoImageNodeFilterStrategy func(*html.Node) bool, imageLinkExtractionStrategy func(*html.Node) string, imageLinkTransformStrategy func(string) string, getChannelIdStrategy func() string, n int) {
	feeds = append(feeds, &Feed{
		Url:                         url,
		ImageNodeFilterStrategy:     &imageNodeFilterStrategy,
		LogoImageNodeFilterStrategy: &logoImageNodeFilterStrategy,
		ImageLinkExtractionStrategy: &imageLinkExtractionStrategy,
		ImageLinkTransformStrategy:  &imageLinkTransformStrategy,
		NumImages:                   n,
		GetChannelIdStrategy:        &getChannelIdStrategy,
	})
}

// GetLatest gets the latest items, up to MAX_ITEMS, that have not been seen during its last call
func GetLatest() map[*Feed][]*gofeed.Item {
	result := map[*Feed][]*gofeed.Item{}
	for _, f := range feeds {
		feed, err := parser.ParseURL(f.Url)
		if err != nil {
			log.Print(err)
			continue
		}
		items := feed.Items
		if len(items) > MAX_ITEMS {
			items = items[:MAX_ITEMS]
		}
		for _, i := range items {
			if !itemInList(seen[f], i) {
				result[f] = append(result[f], i)
			}
		}
	}
	// Update our seen items
	for feed, items := range result {
		// Push the new items to the front of the seen list
		newSeenList := append(items, seen[feed]...)
		if len(newSeenList) > MAX_ITEMS {
			newSeenList = newSeenList[:MAX_ITEMS]
		}
		seen[feed] = newSeenList
	}
	return result
}

// GetImages returns the first n images from the given url page. If a root class is provided, all
// nodes with that class will be searched for images
// url: The url for the page to parse for images
// filterStrategy: Function to narrow down the try to extract links from
// maxLinks: The maximum number of images to search for
// extractionStrategy: Function to extract the image link data as a string from a given node
// transformStrategy: Function to transform the resulting image links before returning
func GetImages(url string, filterStrategy func(*html.Node) bool, maxLinks int, extractionStrategy func(*html.Node) string, transformStrategy func(string) string) ([]string, error) {
	result := []string{}
	client := httpClientWithCookieJar()
	// Load the page and parse the html
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
		return []string{}, fmt.Errorf("getting %s: %v", url, resp.StatusCode)
	}
	doc, err := html.Parse(resp.Body)
	if err != nil {
		return []string{}, fmt.Errorf("parsing %s as HTML: %v", url, err)
	}
	// Extract images by using the provided filter, extraction, and transform strategy
	nodes := []*html.Node{}
	nodes = getNodesByFunc(doc, filterStrategy, nodes)
	for _, n := range nodes {
		link := extractionStrategy(n)
		if len(link) > 0 && len(result) < maxLinks {
			result = append(result, extractionStrategy(n))
		}
	}
	// Transform the result
	for key, val := range result {
		result[key] = transformStrategy(val)
	}
	return result, nil
}

// NoFilterStrategy returns a default node filter strategy that does not filter
func NoFilterStrategy() func(*html.Node) bool {
	return func(*html.Node) bool {
		return true
	}
}

// DefaultFilterStrategy returns a filter to only find nodes that are anchor or image tags
func DefaultFilterStrategy() func(*html.Node) bool {
	return func(n *html.Node) bool {
		if n.Type == html.ElementNode && (n.Data == "a" || n.Data == "img") {
			for _, a := range n.Attr {
				if isImageAttribute(a.Key) && isImageFormat(a.Val) {
					return true
				}
			}
		}
		return false
	}
}

// ParentNodeFilterFunc constructs a filter based on a condition being met for 1 or more of a node's parents
func ParentNodeFilterFunc(condition func(*html.Node) bool) func(*html.Node) bool {
	return func(n *html.Node) bool {
		p := n.Parent
		for p != nil {
			if condition(p) {
				return true
			}
			p = p.Parent
		}
		return false
	}
}

func DefaultTransformStrategy() func(string) string {
	return func(s string) string {
		return s
	}
}

// DefaultExtractionStrategy returns an extractor for image links in known formats
func DefaultExtractionStrategy() func(*html.Node) string {
	return func(n *html.Node) string {
		for _, a := range n.Attr {
			if isImageAttribute(a.Key) && isImageFormat(a.Val) {
				return a.Val
			}
		}
		return ""
	}
}

// FilterByClass constructs a filter strategy that filters by class
func FilterByClass(class string) func(*html.Node) bool {
	cl := class
	return func(node *html.Node) bool {
		for _, a := range node.Attr {
			if a.Key == "class" {
				classes := strings.Split(a.Val, ",")
				for _, c := range classes {
					if c == cl {
						return true
					}
				}
			}
		}
		return false
	}
}

// FilterByAttr constructs a filter strategy that filters by the presence of a specific attribute
func FilterByAttr(key, val string) func(*html.Node) bool {
	return func(n *html.Node) bool {
		for _, a := range n.Attr {
			if a.Key == key && a.Val == val {
				return true
			}
		}
		return false
	}
}
