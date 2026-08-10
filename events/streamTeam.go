package events

import (
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/spf13/viper"
)

func StreamTeamGoLive(s *discordgo.Session, p *discordgo.PresenceUpdate) {
	if p == nil || p.User == nil || p.User.Bot {
		return
	}

	configuredGuildID := viper.GetString("guildId")
	if configuredGuildID != "" && p.GuildID != configuredGuildID {
		return
	}

	streamTeamRoleID := viper.GetString("streamTeamRoleId")
	isLiveRoleID := viper.GetString("isLiveRoleId")
	if streamTeamRoleID == "" || isLiveRoleID == "" {
		log.Println("streamTeamRoleId or isLiveRoleId missing from config")
		return
	}

	member, err := s.GuildMember(p.GuildID, p.User.ID)
	if err != nil {
		log.Printf("Error fetching member %s for stream role sync: %v", p.User.ID, err)
		return
	}

	hasStreamTeamRole := contains(member.Roles, streamTeamRoleID)
	hasLiveRole := contains(member.Roles, isLiveRoleID)
	isLiveOnTwitch := hasTwitchStreamingActivity(p.Activities)

	if hasStreamTeamRole && isLiveOnTwitch && !hasLiveRole {
		if err := s.GuildMemberRoleAdd(p.GuildID, p.User.ID, isLiveRoleID); err != nil {
			log.Printf("Error adding isLiveRoleId to %s: %v", p.User.ID, err)
		}
		return
	}

	if hasLiveRole && (!hasStreamTeamRole || !isLiveOnTwitch) {
		if err := s.GuildMemberRoleRemove(p.GuildID, p.User.ID, isLiveRoleID); err != nil {
			log.Printf("Error removing isLiveRoleId from %s: %v", p.User.ID, err)
		}
	}

}

func hasTwitchStreamingActivity(activities []*discordgo.Activity) bool {
	for _, activity := range activities {
		if activity == nil || activity.Type != discordgo.ActivityTypeStreaming {
			continue
		}

		url := strings.ToLower(activity.URL)
		name := strings.ToLower(activity.Name)
		if strings.Contains(url, "twitch.tv") || strings.Contains(name, "twitch") {
			return true
		}
	}

	return false

}
