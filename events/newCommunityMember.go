package events

import (
	"fmt"
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

	auditLogs, err := s.GuildAuditLog(m.GuildID, "", "", int(discordgo.AuditLogActionMemberRoleUpdate), 5)
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
		log.Printf("No executor ID found for member %s, assuming self-performed action", m.User.ID)
		executorId = m.User.ID
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
		var message string
		if executorId == m.User.ID {
			message = "<@" + m.User.ID + "> has joined the community!\nSay " + greeting + " to them!"
		} else {
			message = "<@" + executorId + "> has welcomed a new member!\nSay " + greeting + " to <@" + m.User.ID + ">!"
		}
		_, err := s.ChannelMessageSend(viper.GetString("communityMemberGeneralChannelId"), message)
		if err != nil {
			log.Println("Error sending welcome message:", err)
		}

		// Create private welcome thread
		// createWelcomeThread(s, m.User, m.GuildID)
	}
}

func createWelcomeThread(s *discordgo.Session, user *discordgo.User, guildID string) {
	channelID := viper.GetString("communityMemberGeneralChannelId")

	// Create private thread
	threadName := fmt.Sprintf("Welcome %s!", user.Username)
	thread, err := s.ThreadStart(channelID, threadName, discordgo.ChannelTypeGuildPrivateThread, 60)
	if err != nil {
		log.Printf("Error creating welcome thread for %s: %v", user.ID, err)
		return
	}

	// Add the new member to the thread
	err = s.ThreadMemberAdd(thread.ID, user.ID)
	if err != nil {
		log.Printf("Error adding member %s to welcome thread: %v", user.ID, err)
	}

	// Send welcome embeds
	// sendWelcomeEmbeds(s, thread.ID, user, guildID)
}

func sendWelcomeEmbeds(s *discordgo.Session, threadID string, user *discordgo.User, guildID string) {
	// Get guild info for server name
	guild, err := s.Guild(guildID)
	if err != nil {
		log.Printf("Error fetching guild info: %v", err)
		guild = &discordgo.Guild{Name: "this server"}
	}

	// Welcome embed
	welcomeEmbed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("Welcome to %s! 🎉", guild.Name),
		Description: fmt.Sprintf("Hey <@%s>! We're excited to have you as part of our community!", user.ID),
		Color:       0x00ff00, // Green
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: user.AvatarURL(""),
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}

	// Server info embed
	infoEmbed := &discordgo.MessageEmbed{
		Title:       "Getting Started 📚",
		Description: "Here are some helpful resources to get you started:",
		Color:       0x0099ff, // Blue
		Fields: []*discordgo.MessageEmbedField{
			{
				Value: "\nWe want you to get the best value you can out of our" +
					"community, and doing so requires you to review some additional" +
					"steps:\n\n" +
					"**Follow channels and categories that interest you**\n" +
					"- Go check out the <id:browse> channel to see what we offer\n" +
					"**Choose Pingable Roles**\n" +
					"- Visit the <#" + viper.GetString("roleSelectionChannelId") + "> channel and choose any roles for which you'd like to receive notifications\n" +
					"**Getting In-Game Invites**\n" +
					"- Use the <#" + viper.GetString("welcomeChannelId") + "> channel to request game invites from our community using the ticket system. There are also ticket types for updating your Discord nickname to match your main's in-game name. This helps us recognize who you are in-game and in discord!",
				Inline: false,
			},
		},
	}
	infoEmbed2 := &discordgo.MessageEmbed{
		Title: "💬 Additional Information",
		Color: 0x0099ff, // Blue
		Fields: []*discordgo.MessageEmbedField{
			{
				Value: "**About Us & Rules**\n" +
					"- You can read about our community and rules in the <#" + viper.GetString("welcomeChannelId") + "> channel. if you haven't already to get a brief introduction to the community, what we are about, and our very simple set of rules to maintain a pleasant and harmonious atmosphere both in Discord and in-game.\n" +
					"**Organized Raid Teams**\n" +
					"- Check out the <#" + viper.GetString("raidTeamsChannelId") + "> channel if you are interested in a raid team, and feel free to inquire to any of the listed contacts. You can also use the <#" + viper.GetString("welcomeChannelId") + "> channel and submit a ticket to get connected with a raid team that fits your schedule and playstyle with some help from our liason team!",
				Inline: false,
			},
		},
	}
	infoEmbed3 := &discordgo.MessageEmbed{
		Title: "🆘 Need Help?",
		Color: 0x0099ff, // Blue
		Fields: []*discordgo.MessageEmbedField{
			{
				Value: "If you have any questions or need assistance, feel free to ask in the server, or here in this thread. One of our friendly <@&" + viper.GetString("welcomeWagonRoleId") + "> members will be happy to assist you!\n\n" +
					"You can leave this thread at any time, or it will auto-archive after a period of inactivity!",
				Inline: false,
			},
		},
	}

	_, err = s.ChannelMessageSend(threadID, fmt.Sprint("<@&"+viper.GetString("roleApproverId")+">"))

	// Send embeds
	_, err = s.ChannelMessageSendEmbed(threadID, welcomeEmbed)
	if err != nil {
		log.Printf("Error sending welcome embed: %v", err)
	}

	_, err = s.ChannelMessageSendEmbed(threadID, infoEmbed)
	if err != nil {
		log.Printf("Error sending info embed: %v", err)
	}
	_, err = s.ChannelMessageSendEmbed(threadID, infoEmbed2)
	if err != nil {
		log.Printf("Error sending info embed 2: %v", err)
	}
	_, err = s.ChannelMessageSendEmbed(threadID, infoEmbed3)
	if err != nil {
		log.Printf("Error sending info embed 3: %v", err)
	}
}
