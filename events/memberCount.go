package events

import (
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/spf13/viper"
)

// Uses cache to set member count as custom status
func SetMemberCount(s *discordgo.Session, guild *discordgo.Guild) {
	roleID := viper.GetString("communityMemberRole")
	count := 0
	for _, key := range memberCache.Keys() {
		if member, ok := memberCache.Get(key); ok {
			if discordMember, valid := member.(*discordgo.Member); valid {
				for _, role := range discordMember.Roles {
					if role == roleID {
						count++
						break // No need to check other roles if we found the community member role
					}
				}
			}
		}
	}
	totalMembers := count
	status := fmt.Sprintf("Community Memebers: %d", totalMembers)
	err := s.UpdateCustomStatus(status)
	if err != nil {
		log.Println("Error setting status:", err)
	}
}
