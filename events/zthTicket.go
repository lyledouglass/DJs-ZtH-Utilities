package events

import (
	"fmt"
	"log"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/spf13/viper"
)

func createTicketEmbed(s *discordgo.Session, m *discordgo.Message, threadId string, userId string) {
	if len(m.Embeds) < 2 {
		log.Println("Missing embeds")
		return
	}
	embedWithInfo := m.Embeds[1]
	var characterName, realm string

	for _, field := range embedWithInfo.Fields {
		log.Printf("Field: %s - %s", field.Name, field.Value)
		switch field.Name {
		case "Character Name":
			characterName = field.Value
		case "Realm or Server":
			realm = field.Value
		}
	}

	if characterName == "" || realm == "" || userId == "" {
		log.Println("Missing character name, realm, or user ID")
		return
	}

	// Send a message to the ticket thread
	ticketEmbed := &discordgo.MessageEmbed{
		Title: "Copy & Paste for Inviters",
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Character",
				Value:  characterName + " - " + realm,
				Inline: false,
			},
			{
				Name:   "Guild Note",
				Value:  "[XFa:" + characterName + "]",
				Inline: false,
			},
			{
				Name:   "Officer Note",
				Value:  "`<@" + userId + ">`",
				Inline: false,
			},
		},
	}

	_, err := s.ChannelMessageSendComplex(threadId, &discordgo.MessageSend{
		Embeds: []*discordgo.MessageEmbed{ticketEmbed},
		Components: []discordgo.MessageComponent{
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.Button{
						Label:    "Ping Inviters",
						Style:    discordgo.PrimaryButton,
						CustomID: "ping_inviters_" + userId,
					},
					discordgo.Button{
						Label:    "Sorry We Missed You",
						Style:    discordgo.PrimaryButton,
						CustomID: "sorry_missed_you_" + userId,
					},
				},
			},
		},
	})
	if err != nil {
		log.Println("Error sending ticket embed:", err)
	}
}

func OnZthTicketCreate(s *discordgo.Session, t *discordgo.ThreadCreate) {
	ticketChannelId := viper.GetString("ticketChannelId")
	fmt.Println("Ticket channel ID:", t.ParentID)
	if t.ParentID == ticketChannelId {
		mu.Lock()
		defer mu.Unlock()

		time.Sleep(2 * time.Second)
		messages, err := s.ChannelMessages(t.ID, 100, "", "", "")
		if err != nil {
			log.Println("Error fetching messages:", err)
			return
		}

		log.Printf("Fetched %d messages from thread %s", len(messages), t.ID)

		var mentionUser *discordgo.User

		for _, msg := range messages {
			//log.Printf("Processing message %s", msg.ID)
			//log.Printf("Message content: %s", msg.Content)
			//log.Printf("Message embeds: %v", msg.Embeds)
			//log.Printf("Message mentions: %v", msg.Mentions)
			//log.Printf("Message author: %s", msg.Author.ID)
			//log.Printf("Message attachements: %v", msg.Attachments)
			if len(msg.Mentions) > 0 {
				mentionUser = msg.Mentions[0]
				log.Printf("Mentioned user: %s", mentionUser.ID)
				break
			}
		}

		if mentionUser == nil {
			log.Println("No mentioned user found")
			return
		}

		if _, exists := processedThreads[t.ID]; exists {
			for _, msg := range messages {
				if len(msg.Embeds) >= 2 {
					log.Println("Creating ticket embed")
					createTicketEmbed(s, msg, t.ID, mentionUser.ID)
					break
				}
			}
		}
		processedThreads[t.ID] = true
	}
}
