package events

import (
	"log"

	"github.com/bwmarrin/discordgo"
	lru "github.com/hashicorp/golang-lru"
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
	var after string
	for {
		members, err := s.GuildMembers(guildId, after, 1000)
		if err != nil {
			log.Printf("Error fetching guild members: %s", err)
		}
		if len(members) == 0 {
			break
		}
		for _, member := range members {
			memberCache.Add(member.User.ID, member)
			log.Printf("Cached member: %s", member.User.ID)
		}
		after = members[len(members)-1].User.ID
	}
	SetMemberCount(s, &discordgo.Guild{ID: guildId})
}
