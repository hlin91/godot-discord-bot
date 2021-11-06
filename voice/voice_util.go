package voice

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/jonas747/dca"
)

const (
	VIDEO_ID          = `v=[[:ascii:]]{11}`
	VIDEO_TITLE       = `"title": ".+?"`
	VIDEO_DESCRIPTION = `"description": ".+?"`
	VIDEO_IMAGE       = `"https://i\.ytimg\.com/vi/[[:alnum:]]{11}/[[:alnum:]]+?.jpg\?[[:alpha:]]+?=.+?"`
)

var vc map[string]*discordgo.VoiceConnection
var lock map[string]chan interface{}
var skip map[string]chan interface{}
var pause map[string]chan interface{}
var signal interface{}

func init() {
	vc = map[string]*discordgo.VoiceConnection{}
	lock = map[string]chan interface{}{}
	skip = map[string]chan interface{}{}
	pause = map[string]chan interface{}{}
}

// Join the voice channel of a guild
func JoinVoice(s *discordgo.Session, guildID, channelID string) error {
	voice, err := s.ChannelVoiceJoin(guildID, channelID, false, true)
	if err != nil || vc == nil {
		return fmt.Errorf("failed to join voice channel %v: %v", channelID, err)
	}
	vc[guildID] = voice
	lock[guildID] = make(chan interface{}, 1)
	skip[guildID] = make(chan interface{}, 1)
	pause[guildID] = make(chan interface{}, 1)
	return nil
}

// Disconnect from the voice channel of a guild
func LeaveVoice(s *discordgo.Session, guildID string) error {
	if vc == nil {
		return fmt.Errorf("tried to disconnect from nil voice channel")
	}
	err := vc[guildID].Disconnect()
	if err != nil {
		return fmt.Errorf("failed to disconnect from voice channel: %v", err)
	}
	vc[guildID] = nil
	lock[guildID] = nil
	skip[guildID] = nil
	return nil
}

// Stream a url to the given voice channel
func StreamUrl(url, guildID string) error {
	re := regexp.MustCompile(VIDEO_ID)
	id := re.FindString(url)[2:]
	ytdl := exec.Command("youtube-dl", "-f", "251", "-o", "-", id)
	outPipe, err := ytdl.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to retrieve stdout pipe: %v", err)
	}
	err = ytdl.Start()
	if err != nil {
		return fmt.Errorf("failed to start ytdl process: %v", err)
	}
	var token interface{}
	lock[guildID] <- token // Lock the mutex
	defer func(ch chan interface{}) {
		<-ch
	}(lock[guildID])
	options := dca.StdEncodeOptions
	options.BufferedFrames = 1000 // Increase the frame buffer to reduce stuttering
	encoder, err := dca.EncodeMem(outPipe, options)
	if err != nil {
		return fmt.Errorf("failed to create dca encoder: %v", err)
	}
	tick := time.NewTicker(20 * time.Millisecond)
	vc[guildID].Speaking(true)
	defer vc[guildID].Speaking(false)
	for frame, err := encoder.OpusFrame(); err == nil; frame, err = encoder.OpusFrame() {
		select {
		case <-skip[guildID]:
			return nil
		case <-pause[guildID]:
			<-pause[guildID]
		default:
			<-tick.C
			vc[guildID].OpusSend <- frame
		}
	}
	return nil
}

func Pause(guildID string) error {
	if pause[guildID] == nil {
		return fmt.Errorf("no song to pause for guild %v", guildID)
	}
	select {
	case pause[guildID] <- signal:
		return nil
	default:
		return fmt.Errorf("pause signal already sent")
	}
}

func Skip(guildID string) error {
	if skip[guildID] == nil {
		return fmt.Errorf("no song to skip for guild %v", guildID)
	}
	select {
	case skip[guildID] <- signal:
		return nil
	default:
		return fmt.Errorf("skip signal already sent")
	}
}

func UrlToEmbed(url string) (*discordgo.MessageEmbed, error) {
	re := regexp.MustCompile(VIDEO_ID)
	id := re.FindString(url)[2:]
	ytdl := exec.Command("youtube-dl", "--skip-download", "--dump-json", id)
	output, err := ytdl.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to execute command: %v", err)
	}
	outString := string(output)
	re = regexp.MustCompile(VIDEO_TITLE)
	title := strings.Trim(re.FindString(outString)[10:], `"`)
	link := `https://www.youtube.com/watch?v=` + id
	re = regexp.MustCompile(VIDEO_DESCRIPTION)
	desc := strings.Trim(re.FindString(outString)[15:], `"`)
	lines := strings.Split(desc, `\n`)
	if len(lines) > 15 {
		lines = lines[:15]
	}
	desc = strings.Join(lines, "\n")
	re = regexp.MustCompile(VIDEO_IMAGE)
	images := re.FindAllString(outString, -1)
	img := strings.Trim(images[len(images)-1], `"`) // Get the highest res image
	return &discordgo.MessageEmbed{
		URL:         link,
		Type:        discordgo.EmbedTypeArticle,
		Title:       title,
		Description: desc,
		Author: &discordgo.MessageEmbedAuthor{
			Name: "Now playing...",
		},
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: img,
		},
		Color: 0xC4302B,
	}, nil
}
