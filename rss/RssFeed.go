package rss

import (
	"log"
	"github.com/mmcdole/gofeed"
)

// RssFeed is a basic ItemProvider constructed from a URL to an existing RSS feed
type RssFeed struct {
	Url string
}

func NewRssFeed(url string) *RssFeed {
	result := RssFeed{
		Url: url,
	}
	return &result
}

func (r RssFeed) items() []*gofeed.Item {
	feed, err := parser.ParseURL(r.Url)
	if err != nil {
		log.Printf("RssFeed.items: failed to parse url '%v'", r.Url)
		return []*gofeed.Item{}
	}
	return feed.Items
}
