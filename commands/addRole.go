package commands

import (
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/spf13/viper"
)

func AddRoleInteractionCreate(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type == discordgo.InteractionApplicationCommand {
		switch i.ApplicationCommandData().Name {
		case "addrole":
			targetUser := i.ApplicationCommandData().Options[0].UserValue(nil)
			role := i.ApplicationCommandData().Options[1].RoleValue(s, i.GuildID)
			user := i.Member.User

			// Fetch the complete user object
			targetMember, err := s.GuildMember(i.GuildID, targetUser.ID)
			if err != nil {
				log.Println("Error fetching target user:", err)
			}

			// Acknowledge the interaction first to avoid timeout
			err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
			})
			if err != nil {
				log.Println("Error acknowledging interaction:", err)
				return
			}

			// Check if the role requires approval
			rolesRequiringApproval := viper.GetStringSlice("rolesRequiringApproval")
			approvalRole := viper.GetString("roleApproverId")
			if contains(rolesRequiringApproval, role.ID) {
				// Send a message to the approver role for approval
				approvalMessage := user.Username + " has requested to add the `@" + role.Name + "` role to " + "`" + targetMember.User.Username + "`. An approver needs to approve this request. " + "<@&" + approvalRole + ">"
				_, err = s.ChannelMessageSendComplex(viper.GetString("accessControlChannelId"), &discordgo.MessageSend{
					Content: approvalMessage,
					Components: []discordgo.MessageComponent{
						discordgo.ActionsRow{
							Components: []discordgo.MessageComponent{
								discordgo.Button{
									Label:    "Approve",
									Style:    discordgo.PrimaryButton,
									CustomID: "approve_role_" + targetUser.ID + "_" + role.ID,
								},
							},
						},
					},
				})
				if err != nil {
					log.Println("Error sending approval message to access channel:", err)
				}
				// Send an ephemeral message to the user indicating that approval is required
				_, err = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
					Content: "Your request to add the role requires approval from an approver.",
					Flags:   discordgo.MessageFlagsEphemeral,
				})
				if err != nil {
					log.Println("Error sending follow-up message:", err)
				}
				return
			}

			// Add the role to the user
			err = s.GuildMemberRoleAdd(i.GuildID, targetUser.ID, role.ID)
			if err != nil {
				log.Println("Error adding role to user:", err)
				_, err = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
					Content: "Failed to add role to user.",
					Flags:   discordgo.MessageFlagsEphemeral,
				})
				if err != nil {
					log.Println("Error sending follow-up message:", err)
				}
				return
			}

			// Format the message
			successMessage := user.Username + " has added the `@" + role.Name + "` to " + "`" + targetMember.User.Username + "`"

			// Send an ephemeral follow-up message indicating success
			_, err = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
				Content: successMessage,
				Flags:   discordgo.MessageFlagsEphemeral,
			})
			if err != nil {
				log.Println("Error sending follow-up message:", err)
			}
			// Send a message to the access channel indicating success
			_, err = s.ChannelMessageSend(viper.GetString("accessControlChannelId"), successMessage)
			if err != nil {
				log.Println("Error sending message to access channel:", err)
			}
			return
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
