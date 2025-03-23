package events

import (
	"log"

	"github.com/bwmarrin/discordgo"
	lru "github.com/hashicorp/golang-lru"
	"github.com/spf13/viper"
)

var (
	auditLogChannelId string
	messageCache      *lru.Cache
	messageToLog      string
	embed             *discordgo.MessageEmbed
)

func init() {
	var err error
	messageCache, err = lru.New(1000)
	if err != nil {
		log.Fatalf("Error creating message cache: %s", err)
	}
}

func OnMemberJoin(s *discordgo.Session, m *discordgo.GuildMemberAdd) {
	auditLogChannelId = viper.GetString("auditLogChannelId")
	// Send a message to the audit log channel
	message := "User <@" + m.User.ID + "> has joined the server"
	_, err := s.ChannelMessageSend(auditLogChannelId, message)
	if err != nil {
		s.ChannelMessageSend(m.GuildID, "Error sending message to audit log channel")
	}
}

func OnMemberLeave(s *discordgo.Session, m *discordgo.GuildMemberRemove) {
	auditLogChannelId = viper.GetString("auditLogChannelId")
	// Send a message to the audit log channel
	message := "User <@" + m.User.ID + "> has left the server"
	_, err := s.ChannelMessageSend(auditLogChannelId, message)
	if err != nil {
		s.ChannelMessageSend(m.GuildID, "Error sending message to audit log channel")
	}
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
