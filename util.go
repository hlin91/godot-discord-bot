package main

import (
	"log"
	"os"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/harvlin/godot/rss"
	"github.com/mmcdole/gofeed"
)

func getGuildID() string {
	return os.Getenv("GuildID")
}

func getChannelID() string {
	return os.Getenv("ChannelID")
}

// Create an embed from a function
func itemToEmbed(item *gofeed.Item, images []string) *discordgo.MessageEmbed {
	logo, _ := rss.GetImages(item.Link, "", 1)
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

func rssProcess(d *discordgo.Session, t *time.Ticker, done chan interface{}, ret chan interface{}) {
	rss.GetLatest() // Initialize the seen posts
	for true {
		select {
		case <-done:
			var signal interface{}
			ret <- signal
			return
		case <-t.C:
			log.Println("checking latest posts...")
			items := rss.GetLatest()
			for source, list := range items {
				for _, i := range list {
					images, _ := rss.GetImages(i.Link, source.Class, source.NumImages)
					d.ChannelMessageSendEmbed(getChannelID(), itemToEmbed(i, images))
				}
			}
		}
	}
}
