package voice

import (
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/jonas747/dca"
)

const (
	VIDEO_ID          = `v=[[:ascii:]]{11}`
	VIDEO_TITLE       = `"title": ".+?"`
	VIDEO_DESCRIPTION = `"description": ".+?"`
	VIDEO_IMAGE       = `"https://i\.ytimg\.com/vi/[[:ascii:]]{11}/[[:alnum:]]+?.jpg\?[[:alpha:]]+?=.+?"`
)

var vc map[string]*discordgo.VoiceConnection
var lock map[string]chan interface{}
var skip map[string]chan interface{}
var pause map[string]chan interface{}
var signal interface{}
var recentlyPlayed map[string]string // Maps song titles to urls

func init() {
	vc = map[string]*discordgo.VoiceConnection{}
	lock = map[string]chan interface{}{}
	skip = map[string]chan interface{}{}
	pause = map[string]chan interface{}{}
	recentlyPlayed = map[string]string{}
}

// Join the voice channel of a guild
func JoinVoice(s *discordgo.Session, guildID, channelID string) error {
	voice, err := s.ChannelVoiceJoin(guildID, channelID, false, true)
	if err != nil || vc[guildID] == nil {
		return fmt.Errorf("failed to join voice channel %v: %v", channelID, err)
	}
	vc[guildID] = voice
	lock[guildID] = make(chan interface{}, 1)
	return nil
}

// Disconnect from the voice channel of a guild
func LeaveVoice(guildID string) error {
	if vc[guildID] == nil {
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
	if vc[guildID] == nil {
		return fmt.Errorf("tried to play into nil voice channel")
	}
	re := regexp.MustCompile(VIDEO_ID)
	id := re.FindString(url)[2:]
	ytdl := exec.Command("youtube-dl", "-f", "251", "-o", "-", id)
	outPipe, err := ytdl.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to retrieve stdout pipe: %v", err)
	}
	options := dca.StdEncodeOptions
	options.BufferedFrames = 1000 // Increase the frame buffer to reduce stuttering
	encoder, err := dca.EncodeMem(outPipe, options)
	if err != nil {
		return fmt.Errorf("failed to create dca encoder: %v", err)
	}
	tick := time.NewTicker(20 * time.Millisecond)
	var token interface{}
	lock[guildID] <- token // Lock the mutex
	skip[guildID] = make(chan interface{}, 1)
	pause[guildID] = make(chan interface{}, 1)
	defer func() {
		pause[guildID] = nil
		skip[guildID] = nil
		<-lock[guildID]
	}()
	vc[guildID].Speaking(true)
	defer vc[guildID].Speaking(false)
	err = ytdl.Start()
	if err != nil {
		return fmt.Errorf("failed to start ytdl process: %v", err)
	}
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
	title, err := strconv.Unquote(re.FindString(outString)[9:])
	if err != nil {
		title = fmt.Sprintf("%v", err)
	}
	link := `https://www.youtube.com/watch?v=` + id
	re = regexp.MustCompile(VIDEO_DESCRIPTION)
	desc, err := strconv.Unquote(re.FindString(outString)[15:])
	if err != nil {
		desc = fmt.Sprintf("%v", err)
	}
	lines := strings.Split(desc, "\n")
	if len(lines) > 20 {
		lines = lines[:20]
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

func streamUrlCoroutine(s *discordgo.Session, i *discordgo.InteractionCreate, url, gID string, info *discordgo.MessageEmbed) {
	var signal interface{}
	StatusLock <- signal
	s.UpdateListeningStatus(info.Title)
	defer func(s *discordgo.Session) {
		s.UpdateListeningStatus("")
		<-StatusLock
	}(s)
	err := StreamUrl(url, gID)
	if err != nil {
		info.Author = &discordgo.MessageEmbedAuthor{
			Name: fmt.Sprintf("Error occurred during playback: \n%v", err),
		}
		s.InteractionResponseEdit(s.State.User.ID, i.Interaction, &discordgo.WebhookEdit{
			Embeds: []*discordgo.MessageEmbed{info},
		})
		return
	}
	info.Author = &discordgo.MessageEmbedAuthor{
		Name: "Finished playing",
	}
	s.InteractionResponseEdit(s.State.User.ID, i.Interaction, &discordgo.WebhookEdit{
		Embeds: []*discordgo.MessageEmbed{info},
	})
}

func getSelectMenuOptionsFromRecentlyPlayed() []discordgo.SelectMenuOption {
	result := []discordgo.SelectMenuOption{}
	for title, url := range recentlyPlayed {
		result = append(result, discordgo.SelectMenuOption{
			Label:       title,
			Value:       url,
			Default:     false,
			Description: "Play this song",
		})
	}
	if len(result) > 25 {
		result = result[:25]
	}
	return result
}
