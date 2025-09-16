package posts

import (
	"log"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/spf13/viper"
)

func EmbedRemover(s *discordgo.Session) {
	// Get channels from config
	embedRemovalChannels := viper.GetStringSlice("embedRemoveChannels")
	if len(embedRemovalChannels) == 0 {
		log.Println("No embed removal channels configured")
		return
	}

	// Create a map for faster channel lookup
	channelMap := make(map[string]bool)
	for _, channel := range embedRemovalChannels {
		channelMap[strings.TrimSpace(channel)] = true
	}

	// Add message event handler
	s.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		// Ignore bot messages
		if m.Author.Bot {
			return
		}

		// Check if message is in a configured channel
		if !channelMap[m.ChannelID] {
			return
		}

		// Check if message contains URLs (potential embeds)
		if strings.Contains(m.Content, "http") {
			// Wait a moment for embeds to load, then suppress them
			go func() {
				time.Sleep(2 * time.Second)

				// Fetch the message again to see if embeds were added
				message, err := s.ChannelMessage(m.ChannelID, m.ID)
				if err != nil {
					return
				}

				// If message has embeds, suppress them
				if len(message.Embeds) > 0 {
					// Suppress embeds using the message flags
					_, err = s.ChannelMessageEditComplex(&discordgo.MessageEdit{
						ID:      m.ID,
						Channel: m.ChannelID,
						Flags:   discordgo.MessageFlagsSuppressEmbeds,
					})
					if err != nil {
						log.Printf("Failed to suppress embeds in channel %s: %v", m.ChannelID, err)
					}
				}
			}()
		}
	})
}
