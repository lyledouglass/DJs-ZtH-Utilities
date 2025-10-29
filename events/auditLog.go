package events

import (
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	lru "github.com/hashicorp/golang-lru"
	"github.com/spf13/viper"
)

var (
	auditLogChannelId string
	messageCache      *lru.Cache
	memberCache       *lru.Cache
	roleCommandCache  *lru.Cache
	messageToLog      string
	embed             *discordgo.MessageEmbed
)

// Initialize the role command cache if needed
func init() {
	if roleCommandCache == nil {
		roleCommandCache, _ = lru.New(1000)
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

// TrackRoleCommand tracks who invoked a role command
func TrackRoleCommand(targetUserID, invokerUserID, roleID string) {
	if roleCommandCache == nil {
		roleCommandCache, _ = lru.New(1000)
	}
	key := targetUserID + ":" + roleID
	roleCommandCache.Add(key, invokerUserID)

	log.Printf("TrackRoleCommand: Tracking key=%s, invoker=%s", key, invokerUserID)

	// Remove the entry after 10 seconds to prevent stale data
	go func() {
		time.Sleep(10 * time.Second)
		roleCommandCache.Remove(key)
		log.Printf("TrackRoleCommand: Removed expired key=%s", key)
	}()
}

// Helper function to get the user who performed the role change
func getResponsibleUser(s *discordgo.Session, guildID string, targetUserID string, actionType discordgo.AuditLogAction) string {
	// First check if we have a tracked role command invoker
	if roleCommandCache != nil {
		log.Printf("getResponsibleUser: Checking cache for targetUserID=%s", targetUserID)
		// Check for any role that might have been added/removed for this user
		for _, key := range roleCommandCache.Keys() {
			keyStr := key.(string)
			log.Printf("getResponsibleUser: Checking cached key=%s", keyStr)
			if strings.HasPrefix(keyStr, targetUserID+":") {
				if invokerID, exists := roleCommandCache.Get(key); exists {
					log.Printf("getResponsibleUser: Found cached invoker=%s for key=%s", invokerID.(string), keyStr)
					return "<@" + invokerID.(string) + ">"
				}
			}
		}
		log.Printf("getResponsibleUser: No cached invoker found for targetUserID=%s", targetUserID)
	}

	// Fall back to audit log check
	auditLogs, err := s.GuildAuditLog(guildID, "", "", int(actionType), 50)
	if err != nil {
		log.Printf("Error fetching audit log: %v", err)
		return "Unknown"
	}

	// Look for the most recent audit log entry for this user
	for _, entry := range auditLogs.AuditLogEntries {
		if entry.TargetID == targetUserID {
			// Convert string ID to int64
			entryIDInt, err := strconv.ParseInt(entry.ID, 10, 64)
			if err != nil {
				log.Printf("Error parsing audit log entry ID: %v", err)
				continue
			}

			// Check if this entry is recent (within last 30 seconds)
			entryTime := time.Unix((entryIDInt>>22)/1000+1420070400, 0)
			if time.Since(entryTime) < 30*time.Second {
				userID := entry.UserID
				user, err := s.User(userID)
				if err == nil && user.Bot {
					return "Bot (via slash command)"
				}
				return "<@" + userID + ">"
			}
			break
		}
	}
	return "Unknown"
}

func OnMemberUpdate(s *discordgo.Session, m *discordgo.GuildMemberUpdate) {
	log.Printf("OnMemberUpdate triggered for user: %s", m.User.ID)

	// Check if roles have changed by comparing with cached member
	rolesChanged := false
	var addedRoles []string
	var removedRoles []string
	cachedMember, exists := memberCache.Get(m.User.ID)
	if exists {
		member, ok := cachedMember.(*discordgo.Member)
		if ok {
			// Compare role slices
			if len(m.Roles) != len(member.Roles) {
				rolesChanged = true
			} else {
				// Check if all roles match
				roleMap := make(map[string]bool)
				for _, role := range member.Roles {
					roleMap[role] = true
				}
				for _, role := range m.Roles {
					if !roleMap[role] {
						rolesChanged = true
						break
					}
				}
			}

			// Find added roles
			oldRoleMap := make(map[string]bool)
			for _, role := range member.Roles {
				oldRoleMap[role] = true
			}
			for _, role := range m.Roles {
				if !oldRoleMap[role] {
					addedRoles = append(addedRoles, role)
				}
			}

			// Find removed roles
			newRoleMap := make(map[string]bool)
			for _, role := range m.Roles {
				newRoleMap[role] = true
			}
			for _, role := range member.Roles {
				if !newRoleMap[role] {
					removedRoles = append(removedRoles, role)
				}
			}
		}
	} else {
		// No cached member, assume roles changed
		rolesChanged = true
		// All current roles are considered "added"
		addedRoles = m.Roles
	}

	// Log role additions to audit channel (excluding open roles)
	if len(addedRoles) > 0 {
		auditLogChannelId = viper.GetString("auditLogChannelId")
		openRoles := viper.GetStringMapString("openRoles")

		// Filter out open roles
		restrictedAddedRoles := []string{}
		for _, role := range addedRoles {
			// Check if this role ID exists in the openRoles map
			if _, isOpenRole := openRoles[role]; !isOpenRole {
				restrictedAddedRoles = append(restrictedAddedRoles, role)
			}
		}

		// Send audit log message if there are restricted role additions
		if len(restrictedAddedRoles) > 0 {
			rolesText := ""
			for _, role := range restrictedAddedRoles {
				rolesText += "<@&" + role + "> "
			}

			responsibleUser := getResponsibleUser(s, m.GuildID, m.User.ID, discordgo.AuditLogActionMemberRoleUpdate)

			embed := &discordgo.MessageEmbed{
				Title: "Role(s) Added",
				Color: 0x00FF00, // Green color for role addition
				Fields: []*discordgo.MessageEmbedField{
					{
						Name:  "User",
						Value: "<@" + m.User.ID + ">",
					},
					{
						Name:  "Roles Added",
						Value: rolesText,
					},
					{
						Name:  "Added By",
						Value: responsibleUser,
					},
				},
			}

			_, err := s.ChannelMessageSendComplex(auditLogChannelId, &discordgo.MessageSend{
				Embeds: []*discordgo.MessageEmbed{embed},
			})
			if err != nil {
				log.Printf("Error sending role addition audit log: %v", err)
			}
		}
	}

	// Log role removals to audit channel (excluding open roles)
	if len(removedRoles) > 0 {
		auditLogChannelId = viper.GetString("auditLogChannelId")
		openRoles := viper.GetStringMapString("openRoles")

		// Filter out open roles
		restrictedRemovedRoles := []string{}
		for _, role := range removedRoles {
			// Check if this role ID exists in the openRoles map
			if _, isOpenRole := openRoles[role]; !isOpenRole {
				restrictedRemovedRoles = append(restrictedRemovedRoles, role)
			}
		}

		// Send audit log message if there are restricted role removals
		if len(restrictedRemovedRoles) > 0 {
			rolesText := ""
			for _, role := range restrictedRemovedRoles {
				rolesText += "<@&" + role + "> "
			}

			responsibleUser := getResponsibleUser(s, m.GuildID, m.User.ID, discordgo.AuditLogActionMemberRoleUpdate)

			embed := &discordgo.MessageEmbed{
				Title: "Role(s) Removed",
				Color: 0xFF8C00, // Orange color for role removal
				Fields: []*discordgo.MessageEmbedField{
					{
						Name:  "User",
						Value: "<@" + m.User.ID + ">",
					},
					{
						Name:  "Roles Removed",
						Value: rolesText,
					},
					{
						Name:  "Removed By",
						Value: responsibleUser,
					},
				},
			}

			_, err := s.ChannelMessageSendComplex(auditLogChannelId, &discordgo.MessageSend{
				Embeds: []*discordgo.MessageEmbed{embed},
			})
			if err != nil {
				log.Printf("Error sending role removal audit log: %v", err)
			}
		}
	}

	// Only call WelcomeNewCommunityMember if roles changed
	if rolesChanged {
		WelcomeNewCommunityMember(s, m)
	}

	// Delay the cache update to allow other handlers (e.g.
	// WelcomeNewCommunityMember) to process first
	delay := viper.GetInt("memberCacheUpdateDelay")
	go func() {
		time.Sleep(time.Duration(delay) * time.Millisecond)
		memberCache.Add(m.User.ID, &discordgo.Member{
			User:  m.User,
			Roles: m.Roles,
		})
		log.Printf("Updated cache for member: %s", m.User.ID)
	}()
}

func OnMemberLeave(s *discordgo.Session, m *discordgo.GuildMemberRemove) {
	rolesToPing := []string{}
	auditLogChannelId = viper.GetString("auditLogChannelId")
	accessControlChannelId := viper.GetString("accessControlChannelId")
	// Send a message to the audit log channel
	message := "User <@" + m.User.ID + "> has left the server"

	// Combine both role lists into restrictedRoles
	rolesRequiringApproval := viper.GetStringSlice("rolesRequiringApproval")
	approvedRoles := viper.GetStringSlice("approvedRoles")
	restrictedRoles := append(rolesRequiringApproval, approvedRoles...)

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
		Color: 0xFF0000, // Red color for user leaving
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
