package events

import (
	"log"

	"github.com/bwmarrin/discordgo"
	lru "github.com/hashicorp/golang-lru"
	"github.com/spf13/viper"
)

func init() {
	var err error
	messageCache, err = lru.New(1000)
	if err != nil {
		log.Fatalf("Error creating message cache: %s", err)
	}
	memberCache, err = lru.New(3000)
	if err != nil {
		log.Fatalf("Error creating member cache: %s", err)
	}
}

func CacheGuildMembers(s *discordgo.Session, guildId string) {
	members, err := s.GuildMembers(guildId, "", 1000)
	if err != nil {
		log.Printf("Error fetching guild members: %s", err)
	}
	for _, member := range members {
		if hasRole(member.Roles, viper.GetString("communityMemberRole")) {
			memberCache.Add(member.User.ID, member)
			log.Printf("Cached member: %s", member.User.ID)
		}
	}
}
