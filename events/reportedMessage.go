package events

import (
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/spf13/viper"
)

func HandleReportMessageCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionApplicationCommand {
		return
	}
	data := i.ApplicationCommandData()
	if data.Name != "Report Message" {
		return
	}
	messageId := data.TargetID
	channelId := i.ChannelID

	moderatorRole := viper.GetString("moderatorRoleId")
	modChannelId := viper.GetString("moderationChannelId")
	messageLink := "https://discord.com/channels/" + i.GuildID + "/" + channelId + "/" + messageId

	embed := &discordgo.MessageEmbed{
		Title:       "Message Reported",
		Description: "Message Reported: " + messageLink + " has been reported.",
	}

	_, err := s.ChannelMessageSendComplex(modChannelId, &discordgo.MessageSend{
		Content: "<@&" + moderatorRole + ">",
		Embeds:  []*discordgo.MessageEmbed{embed},
	})
	if err != nil {
		log.Println("Error sending report message:", err)
	}
	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Message has been reported.",
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
	if err != nil {
		log.Println("Error sending interaction response:", err)
	}
}
