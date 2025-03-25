package commands

import (
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/spf13/viper"
)

func RemoveRole(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type == discordgo.InteractionApplicationCommand {
		switch i.ApplicationCommandData().Name {
		case "removerole":
			// Fetch the user and role from the interaction
			targetUser := i.ApplicationCommandData().Options[0].UserValue(nil)
			role := i.ApplicationCommandData().Options[1].RoleValue(s, i.GuildID)
			user := i.Member.User

			// Fetch the complete user object
			targetMember, err := s.GuildMember(i.GuildID, targetUser.ID)
			if err != nil {
				log.Println("Error fetching target user:", err)
			}

			err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Flags: discordgo.MessageFlagsEphemeral,
				},
			})
			if err != nil {
				log.Println("Error acknowledging interaction:", err)
				return
			}
			if notContains(targetMember.Roles, role.ID) {
				_, err = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
					Content: targetMember.User.Username + " does not have the role.",
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
					Title:       "Role Removal Request",
					Description: user.Username + " has requested to remove the `@" + role.Name + "` role from " + "`" + targetMember.User.Username + "`",
					Color:       0xff0000,
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
									CustomID: "approve_remove_role_" + targetUser.ID + "_" + role.ID,
								},
								discordgo.Button{
									Label:    "Deny",
									Style:    discordgo.DangerButton,
									CustomID: "deny_remove_role_" + targetUser.ID + "_" + role.ID,
								},
							},
						},
					},
				})
				if err != nil {
					log.Println("Error sending message:", err)
				}
				_, err = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
					Content: "Your request to remove the role has been sent for approval.",
					Flags:   discordgo.MessageFlagsEphemeral,
				})
				if err != nil {
					log.Println("Error sending follow-up message:", err)
				}
				return
			}
			// Remove the role
			err = s.GuildMemberRoleRemove(i.GuildID, targetUser.ID, role.ID)
			if err != nil {
				log.Println("Error removing role:", err)
				_, err = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
					Content: "Failed to remove the role.",
					Flags:   discordgo.MessageFlagsEphemeral,
				})
				if err != nil {
					log.Println("Error sending follow-up message:", err)
				}
				return
			}

			executorReturnMessage := "The role `@" + role.Name + "` has been removed from " + "`" + targetMember.User.Username + "`"
			successMessage := "`" + user.Username + "` has removed the role `@" + role.Name + "` from " + "`" + targetMember.User.Username + "`"

			// Send a follow-up message to the user
			_, err = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
				Content: executorReturnMessage,
				Flags:   discordgo.MessageFlagsEphemeral,
			})
			if err != nil {
				log.Println("Error sending follow-up message:", err)
			}
			_, err = s.ChannelMessageSend(viper.GetString("accessControlChannelId"), successMessage)
			if err != nil {
				log.Println("Error sending message:", err)
			}
			return
		}
	}
}

// contains checks if a slice contains a specific string
func notContains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return false
		}
	}
	return true
}
