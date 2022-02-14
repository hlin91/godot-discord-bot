package rss

import (
	"log"
	"strings"
	"testing"

	"github.com/mmcdole/gofeed"
)

func TestRssFeeds(t *testing.T) {
	AddFeed(`http://fiu758.blog111.fc2.com/?xml`, "main_txt", "sh_fc2blogheadbar_body", 1, DefaultImageLinkTransformStrategy())
	AddFeed(`http://2chav.com/?xml`, "kobetu_kiji", "", 1, DefaultImageLinkTransformStrategy())
	AddFeed(`https://dlsite-rss.s3-ap-northeast-1.amazonaws.com/voice_rss.xml`, "work_parts_multitype_item type_contents", "logo", 1, func(s string) string {
		if strings.HasPrefix(s, "//") {
			return `https:` + s
		}
		return `https://www.dlsite.com` + s
	})

	items := GetLatest()
	var item *gofeed.Item
	images := []string{}
	logos := []string{}
	for key, val := range items {
		if len(val) == 0 {
			t.Errorf("got 0 items for %v, want > 0", key)
		}
		item = val[0]
		images, _ = GetImages(item.Link, key.Class, key.NumImages, *key.ImageLinkTransformStrategy)
		logos, _ = GetImages(item.Link, key.LogoClass, key.NumImages, *key.ImageLinkTransformStrategy)
		if len(images) == 0 {
			t.Errorf("got 0 images for %q, want 1", item.Link)
		}
		if len(logos) == 0 {
			t.Errorf("got 0 logos for %q, want 1", item.Link)
		}
		log.Printf("link: %v\nimage: %v\nlogo: %v\n", item.Link, images[0], logos[0])
	}
}
