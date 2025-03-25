package events

import (
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
	lru "github.com/hashicorp/golang-lru"
	"github.com/spf13/viper"
)

var (
	auditLogChannelId string
	messageCache      *lru.Cache
	memberCache       *lru.Cache
	messageToLog      string
	embed             *discordgo.MessageEmbed
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

func OnMemberJoin(s *discordgo.Session, m *discordgo.GuildMemberAdd) {
	auditLogChannelId = viper.GetString("auditLogChannelId")
	// Cache the member
	memberCache.Add(m.User.ID, &discordgo.Member{
		User:  m.User,
		Roles: m.Roles,
	})
	log.Printf("Cached member: %s", m.User.ID)
	// Send a message to the audit log channel
	message := "User <@" + m.User.ID + "> has joined the server"
	_, err := s.ChannelMessageSend(auditLogChannelId, message)
	if err != nil {
		s.ChannelMessageSend(m.GuildID, "Error sending message to audit log channel")
	}
}

func OnMemberUpdate(s *discordgo.Session, m *discordgo.GuildMemberUpdate) {
	communityMemberRole := viper.GetString("communityMemberRole")
	if hasRole(m.Roles, communityMemberRole) {
		// Check if the member is already cached
		memberCache.Add(m.User.ID, &discordgo.Member{
			User:  m.User,
			Roles: m.Roles,
		})
		log.Printf("Cached member: %s", m.User.ID)
	}
}

func OnMemberLeave(s *discordgo.Session, m *discordgo.GuildMemberRemove) {
	rolesToPing := []string{}
	auditLogChannelId = viper.GetString("auditLogChannelId")
	accessControlChannelId := viper.GetString("accessControlChannelId")
	// Send a message to the audit log channel
	message := "User <@" + m.User.ID + "> has left the server"
	restrictedRoles := viper.GetStringSlice("rolesRequiringApproval")

	// All members with restricted roles should have the Community Member
	// role by default if someone hasn't done the process incorrectly, so
	// we can assume they are in the cache
	cachedMember, exists := memberCache.Get(m.User.ID)
	if !exists {
		log.Printf("Member %s not found in cache", m.User.ID)
		return
	}

	member, ok := cachedMember.(*discordgo.Member)
	if !ok {
		log.Printf("Cached member %s is not of type *discordgo.Member", m.User.ID)
		return
	}
	log.Printf("Processing leave event for member %s with cached roles: %v", m.User.ID, member.Roles)

	// Check if the user has any restricted roles
	for _, role := range member.Roles {
		log.Printf("Role: %s", role)
		if contains(restrictedRoles, role) {
			rolesToPing = append(rolesToPing, role)
		}
	}

	embed := &discordgo.MessageEmbed{
		Title: "User Left",
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:  "User",
				Value: "<@" + m.User.ID + ">",
			},
			{
				Name: "Roles",
				Value: "Roles removed: " + strings.Join(func(roles []string) []string {
					formattedRoles := make([]string, len(roles))
					for i, role := range roles {
						formattedRoles[i] = "\n<@&" + role + ">"
					}
					return formattedRoles
				}(member.Roles), ""),
			},
		},
	}

	if len(rolesToPing) > 0 {
		rolesMention := ""
		for _, role := range rolesToPing {
			rolesMention += "<@&" + role + ">"
		}

		_, err := s.ChannelMessageSendComplex(accessControlChannelId, &discordgo.MessageSend{
			Content: rolesMention,
			Embeds:  []*discordgo.MessageEmbed{embed},
		})
		if err != nil {
			s.ChannelMessageSend(m.GuildID, "Error sending message to audit log channel")
		}

	}

	_, err := s.ChannelMessageSend(auditLogChannelId, message)
	if err != nil {
		s.ChannelMessageSend(m.GuildID, "Error sending message to audit log channel")
	}

	// Remove the member from the cache
	memberCache.Remove(m.User.ID)
	log.Printf("Removed member %s from cache", m.User.ID)
}

func OnMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	messageCache.Add(m.ID, m.Message)
}

func OnMessageUpdate(s *discordgo.Session, m *discordgo.MessageUpdate) {
	messageCache.Add(m.ID, m.Message)
}

func OnMessageDelete(s *discordgo.Session, m *discordgo.MessageDelete) {
	auditLogChannelId = viper.GetString("auditLogChannelId")
	deletedMessage, exists := messageCache.Get(m.ID)
	if !exists {
		messageToLog = "Message was not cached, unable to retrieve message content"
		embed = &discordgo.MessageEmbed{
			Title: "Message Deleted",
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:  "Channel",
					Value: "<#" + m.ChannelID + ">",
				},
				{
					Name:  "Author",
					Value: "Unknown",
				},
				{
					Name:  "Message",
					Value: messageToLog,
				},
			},
		}
	} else {
		messageToLog = deletedMessage.(*discordgo.Message).Content
		embed = &discordgo.MessageEmbed{
			Title: "Message Deleted",
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:  "Channel",
					Value: "<#" + m.ChannelID + ">",
				},
				{
					Name:  "Author",
					Value: "<@" + deletedMessage.(*discordgo.Message).Author.ID + ">",
				},
				{
					Name:  "Message",
					Value: messageToLog,
				},
			},
		}
	}

	// Send a message to the audit log channel
	_, err := s.ChannelMessageSendComplex(auditLogChannelId, &discordgo.MessageSend{
		Embeds: []*discordgo.MessageEmbed{embed},
	})
	if err != nil {
		s.ChannelMessageSend(m.GuildID, "Error sending message to audit log channel")
	}
}

// Utility function to check if a user has a specific role
func hasRole(userRoles []string, roleID string) bool {
	for _, role := range userRoles {
		if role == roleID {
			return true
		}
	}
	return false
}
