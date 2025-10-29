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

		// Retry mechanism with increasing delays
		maxRetries := 5
		delays := []time.Duration{2 * time.Second, 3 * time.Second, 5 * time.Second, 8 * time.Second, 10 * time.Second}

		var mentionUser *discordgo.User
		var messageWithEmbeds *discordgo.Message

		for attempt := 0; attempt < maxRetries; attempt++ {
			time.Sleep(delays[attempt])

			messages, err := s.ChannelMessages(t.ID, 100, "", "", "")
			if err != nil {
				log.Printf("Error fetching messages (attempt %d): %v", attempt+1, err)
				continue
			}

			log.Printf("Attempt %d: Fetched %d messages from thread %s", attempt+1, len(messages), t.ID)

			// Find mentioned user if we haven't already
			if mentionUser == nil {
				for _, msg := range messages {
					if msg == nil || msg.Author == nil {
						continue
					}
					if len(msg.Mentions) > 0 && msg.Mentions[0] != nil {
						mentionUser = msg.Mentions[0]
						log.Printf("Mentioned user: %s", mentionUser.ID)
						break
					}
				}
			}

			// Look for message with embeds
			for _, msg := range messages {
				if msg != nil && len(msg.Embeds) >= 2 {
					messageWithEmbeds = msg
					log.Printf("Found message with %d embeds on attempt %d", len(msg.Embeds), attempt+1)
					break
				}
			}

			// If we have both user and embeds, we can proceed
			if mentionUser != nil && messageWithEmbeds != nil {
				log.Println("Creating ticket embed")
				createTicketEmbed(s, messageWithEmbeds, t.ID, mentionUser.ID)
				processedThreads[t.ID] = true
				return
			}

			log.Printf("Attempt %d: mentionUser=%v, messageWithEmbeds=%v", attempt+1, mentionUser != nil, messageWithEmbeds != nil)
		}

		if mentionUser == nil {
			log.Println("No mentioned user found after all retries")
		} else if messageWithEmbeds == nil {
			log.Println("No message with sufficient embeds found after all retries")
		}
	} else {
		log.Printf("Thread created in different channel: %s (not %s)", t.ParentID, ticketChannelId)
	}
}
