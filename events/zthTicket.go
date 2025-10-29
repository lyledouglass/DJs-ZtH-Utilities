package events

import (
	"log"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/spf13/viper"
)

func createTicketEmbed(s *discordgo.Session, m *discordgo.Message, threadId string, userId string) {
	if m == nil || len(m.Embeds) < 2 {
		log.Printf("Message has %d embeds, need at least 2", len(m.Embeds))
		return
	}

	embedWithInfo := m.Embeds[1]
	if embedWithInfo == nil || embedWithInfo.Fields == nil {
		log.Println("Embed or fields are nil")
		return
	}

	log.Printf("Processing embed with %d fields", len(embedWithInfo.Fields))
	var characterName, realm, mainCharacter string

	// Debug: print all fields
	for i, field := range embedWithInfo.Fields {
		if field != nil {
			log.Printf("Field %d: Name='%s', Value='%s'", i, field.Name, field.Value)
		}
	}

	for _, field := range embedWithInfo.Fields {
		switch field.Name {
		case "Character Name":
			characterName = field.Value
		case "Realm or Server":
			realm = field.Value
		case "Main Character":
			mainCharacter = field.Value
		// Add alternative field names in case they're different
		case "Character":
			if characterName == "" {
				characterName = field.Value
			}
		case "Realm", "Server":
			if realm == "" {
				realm = field.Value
			}
		}
	}

	log.Printf("Extracted: Character='%s', Realm='%s', MainCharacter='%s', UserID='%s'", characterName, realm, mainCharacter, userId)

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
				Value:  characterName + "-" + realm,
				Inline: false,
			},
			{
				Name:   "Guild Note",
				Value:  "[XFa:" + mainCharacter + "]",
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
	if t == nil {
		log.Println("ThreadCreate event is nil")
		return
	}

	ticketChannelId := viper.GetString("ticketChannelId")
	log.Printf("Thread created in channel %s, target channel %s", t.ParentID, ticketChannelId)

	if t.ParentID == ticketChannelId {
		mu.Lock()
		defer mu.Unlock()

		// Check if we've already processed this thread
		if _, exists := processedThreads[t.ID]; exists {
			log.Printf("Thread %s already processed, skipping", t.ID)
			return
		}

		log.Printf("Processing new thread %s in ticket channel", t.ID)

		time.Sleep(2 * time.Second)
		messages, err := s.ChannelMessages(t.ID, 100, "", "", "")
		if err != nil {
			log.Printf("Error fetching messages: %v", err)
			return
		}

		log.Printf("Fetched %d messages from thread %s", len(messages), t.ID)

		var mentionUser *discordgo.User

		for _, msg := range messages {
			if msg == nil || msg.Author == nil {
				continue
			}
			log.Printf("Message from %s has %d mentions and %d embeds", msg.Author.Username, len(msg.Mentions), len(msg.Embeds))
			if len(msg.Mentions) > 0 && msg.Mentions[0] != nil {
				mentionUser = msg.Mentions[0]
				log.Printf("Mentioned user: %s", mentionUser.ID)
				break
			}
		}

		if mentionUser == nil {
			log.Println("No mentioned user found")
			return
		}

		// Look for message with embeds and create ticket embed
		for _, msg := range messages {
			if msg != nil && len(msg.Embeds) >= 2 {
				log.Println("Creating ticket embed")
				createTicketEmbed(s, msg, t.ID, mentionUser.ID)
				// Mark as processed only after successful creation
				processedThreads[t.ID] = true
				return
			}
		}

		log.Println("No message with sufficient embeds found")
	} else {
		log.Printf("Thread created in different channel: %s (not %s)", t.ParentID, ticketChannelId)
	}
}
