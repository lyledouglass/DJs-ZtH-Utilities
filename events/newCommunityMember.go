package events

import (
	"log"
	"math/rand"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/spf13/viper"
)

var (
	memberGreetings = []string{
		"Hey there",
		"Hiya",
		"Hello",
		"Hi",
		"Greetings",
		"Bonjour",
		"Howdy",
		"Howdy-do",
		"Heya",
		"Salutations",
		"Oh hi",
		"Hi there",
		"Aloha",
		"Ahoy",
	}
)

func WelcomeNewCommunityMember(s *discordgo.Session, m *discordgo.GuildMemberUpdate) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	greeting := memberGreetings[r.Intn(len(memberGreetings))]

	auditLogs, err := s.GuildAuditLog(m.GuildID, "", "", int(discordgo.AuditLogActionMemberRoleUpdate), 1)
	if err != nil {
		log.Println("Error fetching audit logs:", err)
		return
	}

	var executorId string
	for _, entry := range auditLogs.AuditLogEntries {
		if entry.TargetID == m.User.ID {
			executorId = entry.UserID
			break
		}
	}

	if executorId == "" {
		log.Println("Error finding executor ID")
		return
	}

	communityMemberRole := viper.GetString("communityMemberRole")
	roleAdded := false

	for _, role := range m.Roles {
		if role == communityMemberRole {
			roleAdded = true
			break
		}
	}

	// Retrieve the cached member from the memberCache
	cachedMember, exists := memberCache.Get(m.User.ID)
	if exists {
		member, ok := cachedMember.(*discordgo.Member)
		if !ok {
			log.Printf("Cached member %s is not of type *discordgo.Member", m.User.ID)
			return
		}

		// Check if the role already existed in the cached roles
		for _, role := range member.Roles {
			if role == communityMemberRole {
				// Role already existed before the update, no need to send a message
				roleAdded = false
				break
			}
		}
	} else {
		log.Printf("Member %s not found in cache, assuming role is newly added", m.User.ID)
	}

	// If the role was newly added, send the welcome message
	if roleAdded {
		message := "<@" + executorId + "> has welcomed a new member!\nSay " + greeting + " to <@" + m.User.ID + ">!"
		_, err := s.ChannelMessageSend(viper.GetString("communityMemberGeneralChannelId"), message)
		if err != nil {
			log.Println("Error sending welcome message:", err)
		}
	}
}
