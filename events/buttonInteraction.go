package events

import (
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/spf13/viper"
)

func ButtonInteractionCreate(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type == discordgo.InteractionMessageComponent {
		data := i.MessageComponentData()
		if strings.HasPrefix(data.CustomID, "approve_role_") {
			// Extract targetUserID and roleID from CustomID
			parts := strings.Split(data.CustomID, "_")
			targetUserID := parts[2]
			roleID := parts[3]

			// Check if the user has the approver role
			member, err := s.GuildMember(i.GuildID, i.Member.User.ID)
			if err != nil {
				log.Println("Error fetching member:", err)
				return
			}

			approverRoleId := viper.GetString("roleApproverId")
			if !contains(member.Roles, approverRoleId) {
				err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "You do not have permission to approve this request.",
					},
				})
				if err != nil {
					log.Println("Error sending interaction response:", err)
				}
				return
			}

			// Add the role to the user
			err = s.GuildMemberRoleAdd(i.GuildID, targetUserID, roleID)
			if err != nil {
				log.Println("Error adding role to user:", err)
				err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "Failed to add role to user.",
					},
				})
				if err != nil {
					log.Println("Error sending interaction response:", err)
				}
				return
			}

			// Send a message indicating success
			successMessage := "Role has been successfully added to the user."
			err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: successMessage,
				},
			})
			if err != nil {
				log.Println("Error sending interaction response:", err)
			}
		}
	}
}

// contains checks if a slice contains a specific string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
