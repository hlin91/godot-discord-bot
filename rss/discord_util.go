package rss

import (
	"log"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/mmcdole/gofeed"
)

// ItemToEmbed creates an embed from a function
func ItemToEmbed(item *gofeed.Item, images []string) *discordgo.MessageEmbed {
	logo, _ := GetImages(item.Link, "", 1)
	return &discordgo.MessageEmbed{
		URL:         item.Link,
		Type:        discordgo.EmbedTypeArticle,
		Title:       "<a:newspaper:793257230618460160> **| Hot off the press**\n" + item.Title,
		Description: item.Description,
		Author: &discordgo.MessageEmbedAuthor{
			Name: item.Author.Name,
		},
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: logo[0],
		},
		Image: &discordgo.MessageEmbedImage{
			URL: images[0],
		},
		Color: 0xFCCAEF,
	}
}

// ListenerProcess listens for new RSS posts until a stop signal is sent
func ListenerProcess(d *discordgo.Session, channelID string, t *time.Ticker, done chan interface{}, ret chan interface{}) {
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
					images, _ := GetImages(i.Link, source.Class, source.NumImages)
					d.ChannelMessageSendEmbed(channelID, ItemToEmbed(i, images))
				}
			}
		}
	}
}