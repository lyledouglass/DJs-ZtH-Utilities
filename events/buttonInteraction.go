package events

import (
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/spf13/viper"
)

func RoleButtonInteractionCreate(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type == discordgo.InteractionMessageComponent {
		data := i.MessageComponentData()
		if strings.HasPrefix(data.CustomID, "approve_add_role_") {
			// Extract targetUserID and roleID from CustomID
			parts := strings.Split(data.CustomID, "_")
			targetUserID := parts[3]
			roleID := parts[4]

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

			// Update the embed
			embed := &discordgo.MessageEmbed{
				Title:       "Role Request - Approved",
				Description: "The request to add <@&" + roleID + "> role to <@" + targetUserID + "> has been approved by <@" + i.Member.User.ID + ">.",
				Color:       0x00ff00,
			}

			// Update the message
			_, err = s.ChannelMessageEditComplex(&discordgo.MessageEdit{
				ID:      i.Message.ID,
				Channel: i.ChannelID,
				Embeds:  &[]*discordgo.MessageEmbed{embed},
			})
			if err != nil {
				log.Println("Error editing message:", err)
			}
			// Acknowledge the interaction
			err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseUpdateMessage,
				Data: &discordgo.InteractionResponseData{
					Embeds: []*discordgo.MessageEmbed{embed},
				},
			})
			if err != nil {
				log.Println("Error sending interaction response:", err)
			}
		} else if strings.HasPrefix(data.CustomID, "deny_add_role_") {
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
						Content: "You do not have permission to deny this request.",
					},
				})
				if err != nil {
					log.Println("Error sending interaction response:", err)
				}
				return
			}

			// Update the embed
			embed := &discordgo.MessageEmbed{
				Title:       "Role Request - Denied",
				Description: "The request to add <@&" + roleID + "> to <@" + targetUserID + "> has been denied by <@" + i.Member.User.ID + ">.",
				Color:       0xff0000,
			}

			// Update the message
			_, err = s.ChannelMessageEditComplex(&discordgo.MessageEdit{
				ID:      i.Message.ID,
				Channel: i.ChannelID,
				Embeds:  &[]*discordgo.MessageEmbed{embed},
			})
			if err != nil {
				log.Println("Error editing message:", err)
			}
			// Acknowledge the interaction
			err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseUpdateMessage,
				Data: &discordgo.InteractionResponseData{
					Embeds: []*discordgo.MessageEmbed{embed},
				},
			})
			if err != nil {
				log.Println("Error sending interaction response:", err)
			}
		}
		if strings.HasPrefix(data.CustomID, "approve_remove_role_") {
			parts := strings.Split(data.CustomID, "_")
			targetUserID := parts[3]
			roleID := parts[4]

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

			// Remove the role from the user
			err = s.GuildMemberRoleRemove(i.GuildID, targetUserID, roleID)
			if err != nil {
				log.Println("Error removing role from user:", err)
				err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "Failed to remove role from user.",
					},
				})
				if err != nil {
					log.Println("Error sending interaction response:", err)
				}
				return
			}

			// Update the embed
			embed := &discordgo.MessageEmbed{
				Title:       "Role Removal Request - Approved",
				Description: "The request to remove the <@&" + roleID + "> role from " + "<@" + targetUserID + "> has been approved by <@" + i.Member.User.ID + ">.",
				Color:       0x00ff00,
			}

			// Update the message
			_, err = s.ChannelMessageEditComplex(&discordgo.MessageEdit{
				ID:      i.Message.ID,
				Channel: i.ChannelID,
				Embeds:  &[]*discordgo.MessageEmbed{embed},
			})
			if err != nil {
				log.Println("Error editing message:", err)
			}
			// Acknowledge the interaction
			err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseUpdateMessage,
				Data: &discordgo.InteractionResponseData{
					Embeds: []*discordgo.MessageEmbed{embed},
				},
			})
			if err != nil {
				log.Println("Error sending interaction response:", err)
			}
		} else if strings.HasPrefix(data.CustomID, "deny_remove_role_") {
			if strings.HasPrefix(data.CustomID, "deny_remove_role_") {
				parts := strings.Split(data.CustomID, "_")
				targetUserID := parts[3]
				roleID := parts[4]

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
							Content: "You do not have permission to deny this request.",
						},
					})
					if err != nil {
						log.Println("Error sending interaction response:", err)
					}
					return
				}

				// Update the embed
				embed := &discordgo.MessageEmbed{
					Title:       "Role Removal Request - Denied",
					Description: "The request to remove <@&" + roleID + "> from <@" + targetUserID + "> has been denied by <@" + i.Member.User.ID + ">.",
					Color:       0xff0000,
				}

				// Update the message
				_, err = s.ChannelMessageEditComplex(&discordgo.MessageEdit{
					ID:      i.Message.ID,
					Channel: i.ChannelID,
					Embeds:  &[]*discordgo.MessageEmbed{embed},
				})
				if err != nil {
					log.Println("Error editing message:", err)
				}
				// Acknowledge the interaction
				err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseUpdateMessage,
					Data: &discordgo.InteractionResponseData{
						Embeds: []*discordgo.MessageEmbed{embed},
					},
				})
				if err != nil {
					log.Println("Error sending interaction response:", err)
				}
			}
		}
		if strings.HasPrefix(data.CustomID, "ping_inviters_") {
			parts := strings.Split(data.CustomID, "_")
			userID := parts[2]

			err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseDeferredMessageUpdate,
			})
			if err != nil {
				log.Println("Error sending interaction response:", err)
			}

			s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
				Content: "<@&" + viper.GetString("championRoleId") + ">, our friend " + "<@" + userID + "> is currently awaiting a guild invite!",
				AllowedMentions: &discordgo.MessageAllowedMentions{
					Roles: []string{viper.GetString("championRoleId")},
				},
			})
		} else if strings.HasPrefix(data.CustomID, "sorry_missed_you_") {
			parts := strings.Split(data.CustomID, "_")
			userID := parts[3]

			err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseDeferredMessageUpdate,
			})
			if err != nil {
				log.Println("Error sending interaction response:", err)
			}

			embed := &discordgo.MessageEmbed{
				Title:       "Sorry We Missed You",
				Description: "Sometimes our schedules just don't line up. When you're back online and available for an invite, please hit the Ping Inviters button or mention <@&" + viper.GetString("championRoleId") + "> in this channel and hopefully someone will be available to assist!",
			}

			s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
				Content: "<@" + userID + ">",
				Embeds:  []*discordgo.MessageEmbed{embed},
			})
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
