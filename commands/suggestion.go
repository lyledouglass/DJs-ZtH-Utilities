package commands

import (
	"github.com/bwmarrin/discordgo"
	"github.com/spf13/viper"
)

func Suggestion(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type == discordgo.InteractionApplicationCommand {
		switch i.ApplicationCommandData().Name {
		case "suggestion":
			suggestion := i.ApplicationCommandData().Options[0].StringValue()
			targetChannelName := i.ApplicationCommandData().Options[1].StringValue()
			user := i.Member.User

			// Acknowledge the interaction first to avoid timeout
			err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Flags: discordgo.MessageFlagsEphemeral,
				},
			})
			if err != nil {
				return
			}

			// Send the suggestion to a specific channel (e.g., a suggestions channel)
			leadershipChannels := viper.Get("leadershipChannelIds").([]interface{})

			var targetChannelId string
			for _, channel := range leadershipChannels {
				channelMap := channel.(map[string]interface{})
				if channelMap["name"] == targetChannelName {
					targetChannelId = channelMap["id"].(string)
					targetChannelName, _ = getChannelName(s, targetChannelId)
					break
				}
			}

			if targetChannelId == "" {
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "Channel not found. Please check the channel name.",
						Flags:   discordgo.MessageFlagsEphemeral,
					},
				})
				return
			}

			embed := &discordgo.MessageEmbed{
				Title:       "New Suggestion",
				Description: suggestion,
				Color:       0x00ff00,
				Footer: &discordgo.MessageEmbedFooter{
					Text: "Suggested by " + user.Username,
				},
			}

			_, err = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
				Content: "Your suggestion has been sent to the " + targetChannelName + " team.",
				Flags:   discordgo.MessageFlagsEphemeral,
			})
			if err != nil {
				return
			}

			_, err = s.ChannelMessageSendEmbed(targetChannelId, embed)
			if err != nil {
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "Error sending suggestion. Please try again later.",
						Flags:   discordgo.MessageFlagsEphemeral,
					},
				})
				return
			}
		}
	}
}

func getChannelName(s *discordgo.Session, channelId string) (string, error) {
	channel, err := s.State.Channel(channelId)
	if err != nil {
		return "", err
	}
	return channel.Name, nil
}
