package rss

import (
	"fmt"
	"log"
	"net/http"

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
	Url                        string
	Class                      string
	LogoClass                  string
	NumImages                  int
	ImageLinkTransformStrategy *func(string) string
}

var feeds []Feed                 // List of feeds to parse
var seen map[Feed][]*gofeed.Item // Remember the items we have already seen
var parser *gofeed.Parser

var commands []*discordgo.ApplicationCommand
var commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){}

func init() {
	feeds = []Feed{}
	seen = map[Feed][]*gofeed.Item{}
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
				s.FollowupMessageCreate(s.State.User.ID, i.Interaction, true, &discordgo.WebhookParams{
					Content: "Something went wrong",
				})
				return
			}
			// Do RSS stuff here
			ClearHistory()
			items := GetLatest()
			images := []string{}
			logos := []string{}
			embeds := []*discordgo.MessageEmbed{}
			for key, val := range items {
				images, _ = GetImages(val[0].Link, key.Class, key.NumImages, *key.ImageLinkTransformStrategy)
				logos, _ = GetImages(val[0].Link, key.LogoClass, key.NumImages, *key.ImageLinkTransformStrategy)
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
	return module.CreateModule(commands, commandHandlers)
}

// ClearHistory clears the recently seen lists
func ClearHistory() {
	for key := range seen {
		seen[key] = []*gofeed.Item{}
	}
}

// AddFeed adds a url to the list of feeds to parse
func AddFeed(url string, class string, logo string, n int, imageLinkTransformStrategy func(string) string) {
	feeds = append(feeds, Feed{
		Url:                        url,
		Class:                      class,
		LogoClass:                  logo,
		NumImages:                  n,
		ImageLinkTransformStrategy: &imageLinkTransformStrategy,
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
		seen[feed] = []*gofeed.Item{}
		seen[feed] = append(seen[feed], result[feed]...)
		seen[feed] = append(seen[feed], items[:(len(items)-len(result))]...)
	}
	return result
}

// GetImages returns the first n images from the given url page. If a root class is provided, all
// nodes with that class will be searched for images
// url: The url for the page to parse for images
// class: The root class to parse for images in
// n: The maximum number of images to search for
func GetImages(url, class string, n int, transformStrategy func(string) string) ([]string, error) {
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
	// No class is given so search all anchor and image tags
	if class == "" {
		result = getImagesHelp(doc, "a", result, n)
		result = getImagesHelp(doc, "img", result, n)
		// Transform the result
		for key, val := range result {
			result[key] = transformStrategy(val)
		}
		return result, nil
	}
	// Search all anchor and image tags nested within the given node class
	nodes := []*html.Node{}
	nodes = getNodesByClass(doc, class, nodes)
	for _, node := range nodes {
		result = getImagesHelp(node, "a", result, n)
	}
	for _, node := range nodes {
		result = getImagesHelp(node, "img", result, n)
	}
	// Transform the result
	for key, val := range result {
		result[key] = transformStrategy(val)
	}
	return result, nil
}
