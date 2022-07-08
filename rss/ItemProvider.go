package rss

import (
	"github.com/mmcdole/gofeed"
)

type ItemProvider interface {
	items() []*gofeed.Item
}
