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

			// Acknowledge the interaction first to avoid timeout
			err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Flags: discordgo.MessageFlagsEphemeral,
				},
			})
			if err != nil {
				log.Println("Error acknowledging interaction:", err)
				return
			}

			// Verify that the user has permission to add roles
			if !CheckApprovedRole(s, i.Member) {
				_, err = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
					Content: "You do not have permission to add roles.",
					Flags:   discordgo.MessageFlagsEphemeral,
				})
				if err != nil {
					log.Println("Error sending follow-up message:", err)
				}
				break
			}

			// Fetch the complete user object
			targetMember, err := s.GuildMember(i.GuildID, targetUser.ID)
			if err != nil {
				log.Println("Error fetching target user:", err)
			}

			// Check if the user has the role already
			if contains(targetMember.Roles, role.ID) {
				_, err = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
					Content: targetMember.User.Username + " already has the role.",
					Flags:   discordgo.MessageFlagsEphemeral,
				})
				if err != nil {
					log.Println("Error sending follow-up message:", err)
				}
				return
			}

			// Check if the role requires approval
			rolesRequiringApproval := viper.GetStringSlice("rolesRequiringApproval")
			approvalRole := viper.GetString("roleApproverId")
			if contains(rolesRequiringApproval, role.ID) {
				embed := &discordgo.MessageEmbed{
					Title:       "Role Request",
					Description: "<@" + user.ID + "> has requested to add the <@&" + role.ID + "> role to " + "<@" + targetMember.User.ID + ">",
					Color:       0x00ff00,
					Fields: []*discordgo.MessageEmbedField{
						{
							Name:   "Approval Role",
							Value:  "<@&" + approvalRole + ">",
							Inline: true,
						},
					},
				}
				_, err = s.ChannelMessageSendComplex(viper.GetString("accessControlChannelId"), &discordgo.MessageSend{
					Content: "||<@&" + approvalRole + ">||",
					Embeds:  []*discordgo.MessageEmbed{embed},
					Components: []discordgo.MessageComponent{
						discordgo.ActionsRow{
							Components: []discordgo.MessageComponent{
								discordgo.Button{
									Label:    "Approve",
									Style:    discordgo.PrimaryButton,
									CustomID: "approve_add_role_" + targetMember.User.ID + "_" + role.ID,
								},
								discordgo.Button{
									Label:    "Deny",
									Style:    discordgo.DangerButton,
									CustomID: "deny_add_role_" + targetMember.User.ID + "_" + role.ID,
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
			successEmbed := &discordgo.MessageEmbed{
				Title: "Role Added",
				Color: 0x00ff00,
				Fields: []*discordgo.MessageEmbedField{
					{
						Name:   "Added By",
						Value:  "<@" + user.ID + ">",
						Inline: false,
					},
					{
						Name:   "Target User",
						Value:  "<@" + targetUser.ID + ">",
						Inline: false,
					},
					{
						Name:   "Role",
						Value:  "<@&" + role.ID + ">",
						Inline: false,
					},
				},
			}

			executorReturnMessage := "The `@" + role.Name + "` role has been given to " + "<@" + targetMember.User.ID + ">"

			// Send an ephemeral follow-up message indicating success
			_, err = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
				Content: executorReturnMessage,
				Flags:   discordgo.MessageFlagsEphemeral,
			})
			if err != nil {
				log.Println("Error sending follow-up message:", err)
			}
			// Send a message to the access channel indicating success
			_, err = s.ChannelMessageSendComplex(viper.GetString("accessControlChannelId"), &discordgo.MessageSend{
				Embeds: []*discordgo.MessageEmbed{successEmbed},
			})
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

func CheckApprovedRole(s *discordgo.Session, m *discordgo.Member) (approved bool) {
	approvedRoles := viper.GetStringSlice("rolesRequiringApproval")
	userRoles := m.Roles
	for _, role := range userRoles {
		if contains(approvedRoles, role) {
			return true
		}
	}
	return false
}
