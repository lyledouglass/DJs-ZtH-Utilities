package events

import (
	"log"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/spf13/viper"
)

func createTicketEmbed(s *discordgo.Session, m *discordgo.Message, threadId string, userId string) {
	// Check if thread name starts with "zth-"
	channel, err := s.Channel(threadId)
	if err != nil {
		log.Printf("Error getting channel info: %v", err)
		return
	}

	if !strings.HasPrefix(strings.ToLower(channel.Name), "zth-") {
		log.Printf("Thread name '%s' does not start with 'zth-', skipping embed creation", channel.Name)
		return
	}

	log.Printf("Processing zth ticket: %s", channel.Name)

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

	// Extract Discord nickname from thread messages
	var discordNickname string
	var guildID string

	// Get guild ID from the thread channel itself
	if channel, err := s.Channel(threadId); err == nil && channel.GuildID != "" {
		guildID = channel.GuildID
		log.Printf("Found guild ID from thread channel: %s", guildID)
	} else {
		log.Printf("Error getting thread channel or no guild ID: %v", err)
	}

	messages, err := s.ChannelMessages(threadId, 50, "", "", "")
	if err == nil {
		// If we still don't have guild ID, try to find it from messages as backup
		if guildID == "" {
			for _, msg := range messages {
				if msg != nil && msg.GuildID != "" {
					guildID = msg.GuildID
					log.Printf("Found guild ID from message: %s", guildID)
					break
				}
			}
		}

		for _, msg := range messages {
			if msg == nil || msg.Author == nil {
				continue
			}

			// Look for "Tickets v2 added" system message with more flexible parsing
			if msg.Author.Bot && strings.Contains(msg.Content, "Tickets v2 added") && strings.Contains(msg.Content, "to the thread") {
				// Parse various patterns: "Tickets v2 added <name> to the thread." or similar
				content := msg.Content
				if startIdx := strings.Index(content, "Tickets v2 added "); startIdx != -1 {
					start := startIdx + len("Tickets v2 added ")
					if endIdx := strings.Index(content[start:], " to the thread"); endIdx != -1 {
						discordNickname = strings.TrimSpace(content[start : start+endIdx])
						log.Printf("Extracted Discord nickname from system message: '%s'", discordNickname)
						break
					}
				}
			}

			// Fallback: look for any message from the user to get their display name
			if discordNickname == "" && msg.Author.ID == userId {
				// Try to get nickname from guild member first
				if guildID != "" {
					if member, err := s.GuildMember(guildID, userId); err == nil {
						log.Printf("Guild member lookup from user message - Nick: '%s', Username: '%s', GlobalName: '%s'",
							member.Nick, member.User.Username, member.User.GlobalName)

						if member.Nick != "" && strings.TrimSpace(member.Nick) != "" {
							discordNickname = strings.TrimSpace(member.Nick)
							log.Printf("Extracted Discord nickname from guild member: '%s'", discordNickname)
							break
						} else if member.User.GlobalName != "" && strings.TrimSpace(member.User.GlobalName) != "" {
							discordNickname = strings.TrimSpace(member.User.GlobalName)
							log.Printf("Using Discord global display name from member: '%s'", discordNickname)
							break
						} else {
							discordNickname = member.User.Username
							log.Printf("Using Discord username from member: '%s'", discordNickname)
							break
						}
					} else {
						log.Printf("Error getting guild member from user message: %v", err)
					}
				}
			}

			// Another fallback: look for mentions in embed messages
			if discordNickname == "" && len(msg.Mentions) > 0 {
				for _, mention := range msg.Mentions {
					if mention.ID == userId {
						// Try guild member lookup with the guild ID we found
						if guildID != "" {
							if member, err := s.GuildMember(guildID, userId); err == nil {
								log.Printf("Guild member lookup from mention - Nick: '%s', Username: '%s', GlobalName: '%s'",
									member.Nick, member.User.Username, member.User.GlobalName)

								if member.Nick != "" && strings.TrimSpace(member.Nick) != "" {
									discordNickname = strings.TrimSpace(member.Nick)
									log.Printf("Extracted Discord nickname from mentioned user: '%s'", discordNickname)
								} else if member.User.GlobalName != "" && strings.TrimSpace(member.User.GlobalName) != "" {
									discordNickname = strings.TrimSpace(member.User.GlobalName)
									log.Printf("Using Discord global display name from mention: '%s'", discordNickname)
								} else {
									discordNickname = member.User.Username
									log.Printf("Using Discord username from mention: '%s'", discordNickname)
								}
							} else {
								log.Printf("Error getting guild member from mention: %v", err)
								discordNickname = mention.Username
								log.Printf("Using Discord username from mention (fallback): '%s'", discordNickname)
							}
						} else {
							discordNickname = mention.Username
							log.Printf("Using Discord username from mention (no guild): '%s'", discordNickname)
						}
						break
					}
				}
				if discordNickname != "" {
					break
				}
			}
		}
	} else {
		log.Printf("Error fetching messages for nickname extraction: %v", err)
	}

	// Final fallback: try direct guild member lookup if we still don't have a nickname
	if discordNickname == "" && guildID != "" {
		if member, err := s.GuildMember(guildID, userId); err == nil {
			log.Printf("Guild member lookup successful (final fallback) - Nick: '%s', Username: '%s', GlobalName: '%s'",
				member.Nick, member.User.Username, member.User.GlobalName)

			// Check for nickname first (server-level nickname)
			if member.Nick != "" && strings.TrimSpace(member.Nick) != "" {
				discordNickname = strings.TrimSpace(member.Nick)
				log.Printf("Extracted Discord nickname from direct guild lookup: '%s'", discordNickname)
			} else if member.User.GlobalName != "" && strings.TrimSpace(member.User.GlobalName) != "" {
				// Try global display name (newer Discord feature)
				discordNickname = strings.TrimSpace(member.User.GlobalName)
				log.Printf("Using Discord global display name from direct guild lookup: '%s'", discordNickname)
			} else {
				discordNickname = member.User.Username
				log.Printf("Using Discord username from direct guild lookup: '%s'", discordNickname)
			}
		} else {
			log.Printf("Failed to get guild member for nickname: %v", err)
		}
	}

	// Check if user has a server nickname specifically
	var hasServerNickname bool
	if guildID != "" {
		if member, err := s.GuildMember(guildID, userId); err == nil {
			hasServerNickname = member.Nick != "" && strings.TrimSpace(member.Nick) != ""
		}
	}

	// Use main character as fallback if no Discord nickname found
	guildNoteValue := "[XFa:" + mainCharacter + "]"
	if discordNickname != "" {
		guildNoteValue = "[XFa:" + discordNickname + "]"
	}

	// Send warning embed if user doesn't have a server nickname
	if !hasServerNickname {
		log.Println("Warning: User does not have a server nickname set")

		roleApproverId := viper.GetString("roleApproverId")
		if roleApproverId != "" {
			warningEmbed := &discordgo.MessageEmbed{
				Title:       "⚠️ Missing Server Nickname",
				Description: "User does not have a server nickname set",
				Color:       0xFFA500, // Orange color for warning
				Timestamp:   time.Now().Format(time.RFC3339),
			}

			_, err := s.ChannelMessageSendComplex(threadId, &discordgo.MessageSend{
				Content: "||<@&" + roleApproverId + ">||",
				Embeds:  []*discordgo.MessageEmbed{warningEmbed},
			})
			if err != nil {
				log.Printf("Error sending nickname warning embed: %v", err)
			} else {
				log.Println("Successfully sent server nickname warning embed")
			}
		}
	}

	log.Printf("Extracted: Character='%s', Realm='%s', MainCharacter='%s', DiscordNickname='%s', UserID='%s'", characterName, realm, mainCharacter, discordNickname, userId)

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
				Value:  guildNoteValue,
				Inline: false,
			},
			{
				Name:   "Officer Note",
				Value:  "`<@" + userId + ">`",
				Inline: false,
			},
		},
	}

	_, err = s.ChannelMessageSendComplex(threadId, &discordgo.MessageSend{
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
