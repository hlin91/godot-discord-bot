package rss

import (
	"log"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/mmcdole/gofeed"
)

// ItemToEmbed creates an embed from an Item
func ItemToEmbed(item *gofeed.Item, images []string, logos []string) *discordgo.MessageEmbed {
	if len(logos) == 0 {
		logos = append(logos, "")
	}
	if len(images) == 0 {
		images = append(images, "")
	}
	author := ""
	datePublished := ""
	if item.Author != nil {
		author = item.Author.Name
	}
	if item.PublishedParsed != nil {
		datePublished = item.PublishedParsed.Format(time.ANSIC)
	}
	return &discordgo.MessageEmbed{
		URL:         item.Link,
		Type:        discordgo.EmbedTypeArticle,
		Title:       "<a:newspaper:793257230618460160> **| Hot off the press**\n\n" + item.Title,
		Description: datePublished + "\n\n" + item.Description,
		Author: &discordgo.MessageEmbedAuthor{
			Name: author,
		},
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: logos[0],
		},
		Image: &discordgo.MessageEmbedImage{
			URL: images[0],
		},
		Color: 0xFCCAEF,
	}
}

// ListenerProcess listens for new RSS posts until a stop signal is sent
func ListenerProcess(d *discordgo.Session, t *time.Ticker, done chan interface{}, ret chan interface{}) {
	GetLatest() // Initialize the seen posts
	for {
		select {
		case <-done:
			var signal interface{}
			ret <- signal
			return
		case <-t.C:
			log.Println("checking latest posts...")
			items := GetLatest()
			for source, list := range items {
				for _, i := range list {
					channelId := *source.GetChannelIdStrategy
					images, _ := GetImages(i.Link, *source.ImageNodeFilterStrategy, source.NumImages, *source.ImageLinkExtractionStrategy, *source.ImageLinkTransformStrategy)
					logos, _ := GetImages(i.Link, *source.LogoImageNodeFilterStrategy, source.NumImages, *source.ImageLinkExtractionStrategy, *source.ImageLinkTransformStrategy)
					d.ChannelMessageSendEmbed(channelId(), ItemToEmbed(i, images, logos))
				}
			}
		}
	}
}
