package events

import (
	"log"
	"sync"

	"github.com/bwmarrin/discordgo"
	"github.com/spf13/viper"
)

var processedThreads = make(map[string]bool)
var mu sync.Mutex

func newDJsAppPing(s *discordgo.Session, thread *discordgo.Channel) {
	// When a new app is submitted to the DJs App Forum channel, the bot
	// will ping the DJs Member Role in the new forum post
	roleId := viper.GetString("djsMemberRoleId")
	content := "<@&" + roleId + ">"

	_, err := s.ChannelMessageSend(thread.ID, content)
	if err != nil {
		log.Println("Error sending ping message:", err)
	}
}

func pinDJAppEmbed(s *discordgo.Session, threadId string) {
	messages, err := s.ChannelMessages(threadId, 1, "", "", "")
	if err != nil {
		log.Println("Error fetching messages:", err)
		return
	}
	if len(messages) > 0 {
		err = s.ChannelMessagePin(threadId, messages[0].ID)
		if err != nil {
			log.Println("Error pinning message:", err)
		}
	}
}

func OnDJsThreadCreate(s *discordgo.Session, t *discordgo.ThreadCreate) {
	djsChannelId := viper.GetString("djsAppForumChannelId")
	if t.ParentID == djsChannelId {
		mu.Lock()
		defer mu.Unlock()

		if _, exists := processedThreads[t.ID]; !exists {
			// Run pinDJAppEmbed first to make sure it pins the embed
			pinDJAppEmbed(s, t.ID)
			newDJsAppPing(s, t.Channel)
			processedThreads[t.ID] = true
		}
	}
}
