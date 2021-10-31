package voice

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

var vc map[string]*discordgo.VoiceConnection

func init() {
	vc = map[string]*discordgo.VoiceConnection{}
}

func JoinVoice(s *discordgo.Session, guildID, channelID string) error {
	voice, err := s.ChannelVoiceJoin(guildID, channelID, false, true)
	if err != nil || vc == nil {
		return fmt.Errorf("failed to join voice channel %v: %v", channelID, err)
	}
	vc[guildID] = voice
	return nil
}

func LeaveVoice(s *discordgo.Session, guildID string) error {
	if vc == nil {
		return fmt.Errorf("tried to disconnect from nil voice channel")
	}
	err := vc[guildID].Disconnect()
	if err != nil {
		return fmt.Errorf("failed to disconnect from voice channel: %v", err)
	}
	vc[guildID] = nil
	return nil
}
