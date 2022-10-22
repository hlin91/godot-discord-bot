package rss

import (
	"log"

	"github.com/mmcdole/gofeed"
)

// RssFeed is a basic ItemProvider constructed from a URL to an existing RSS feed
type RssFeed struct {
	Url string
}

func NewRssFeed(url string) *ItemProvider {
	var result ItemProvider = RssFeed{
		Url: url,
	}
	return &result
}

func (r RssFeed) items() []*gofeed.Item {
	feed, err := parser.ParseURL(r.Url)
	if err != nil {
		log.Printf("RssFeed.items: failed to parse url '%v': '%v'", r.Url, err)
		return []*gofeed.Item{}
	}
	return feed.Items
}
