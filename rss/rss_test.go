package rss

import (
	"log"
	"strings"
	"testing"

	"github.com/mmcdole/gofeed"
	"golang.org/x/net/html"
)

func TestRssFeeds(t *testing.T) {
	AddFeed(NewFeedBuilder().WithItemProvider(NewRssFeed(`http://fiu758.blog111.fc2.com/?xml`)).WithImageNodeFilterStrategy(func(n *html.Node) bool {
		parentFilter := ParentNodeFilterFunc(FilterByClass("main_txt"))
		nodeFilter := DefaultFilterStrategy()
		return parentFilter(n) && nodeFilter(n)
	}).WithLogoImageNodeFilterStrategy(func(n *html.Node) bool {
		parentFilter := ParentNodeFilterFunc(FilterByClass("sh_fc2blogheadbar_body"))
		nodeFilter := DefaultFilterStrategy()
		return parentFilter(n) && nodeFilter(n)
	}).Build())

	AddFeed(NewFeedBuilder().WithItemProvider(NewRssFeed(`http://2chav.com/?xml`)).WithImageNodeFilterStrategy(func(n *html.Node) bool {
		parentFilter := ParentNodeFilterFunc(FilterByClass("kobetu_kiji"))
		nodeFilter := DefaultFilterStrategy()
		return parentFilter(n) && nodeFilter(n)
	}).Build())

	AddFeed(NewFeedBuilder().WithItemProvider(NewRssFeed(`https://dlsite-rss.s3-ap-northeast-1.amazonaws.com/voice_rss.xml`)).WithImageNodeFilterStrategy(FilterByAttr("property", "og:image")).WithLogoImageNodeFilterStrategy(func(n *html.Node) bool {
		parentFilter := ParentNodeFilterFunc(FilterByClass("logo"))
		nodeFilter := DefaultFilterStrategy()
		return parentFilter(n) && nodeFilter(n)
	}).WithImageLinkTransformStrategy(func(s string) string {
		if strings.HasPrefix(s, "/") {
			return `https://www.dlsite.com` + s
		}
		return s
	}).Build())

	AddFeed(NewFeedBuilder().WithItemProvider(NewRssFeed(`http://avohayo.blog.fc2.com/?xml`)).WithImageNodeFilterStrategy(func(n *html.Node) bool {
		parentFilter := ParentNodeFilterFunc(FilterByClass("entry_body"))
		nodeFilter := DefaultFilterStrategy()
		return parentFilter(n) && nodeFilter(n)
	}).WithLogoImageNodeFilterStrategy(func(n *html.Node) bool {
		parentFilter := ParentNodeFilterFunc(FilterByAttr("id", "sh_fc2blogheadbar_menu"))
		nodeFilter := DefaultFilterStrategy()
		return parentFilter(n) && nodeFilter(n)
	}).Build())

	items := GetLatest()
	var item *gofeed.Item
	images := []string{}
	logos := []string{}
	for key, val := range items {
		if len(val) == 0 {
			t.Errorf("got 0 items for %v, want > 0", key)
		}
		item = val[0]
		images, _ = GetImages(item.Link, *key.ImageNodeFilterStrategy, key.NumImages, *key.ImageLinkExtractionStrategy, *key.ImageLinkTransformStrategy)
		logos, _ = GetImages(item.Link, *key.LogoImageNodeFilterStrategy, key.NumImages, *key.ImageLinkExtractionStrategy, *key.ImageLinkTransformStrategy)
		if len(images) == 0 {
			t.Errorf("got 0 images for %q, want 1", item.Link)
		}
		if len(logos) == 0 {
			t.Errorf("got 0 logos for %q, want 1", item.Link)
		}
		log.Printf("link: %v\nimages: %v\nlogos: %v\n", item.Link, images, logos)
	}
}
