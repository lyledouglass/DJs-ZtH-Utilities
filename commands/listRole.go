package commands

import (
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func ListRole(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type == discordgo.InteractionApplicationCommand {
		switch i.ApplicationCommandData().Name {
		case "listrole":
			// Fetch the role from the interaction
			role := i.ApplicationCommandData().Options[0].RoleValue(s, i.GuildID)

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

			if !CheckApprovedRole(s, i.Member) {
				_, err = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
					Content: "You do not have permission to list members with roles.",
					Flags:   discordgo.MessageFlagsEphemeral,
				})
				if err != nil {
					log.Println("Error sending follow-up message:", err)
				}
				return
			}

			var allMembers []*discordgo.Member
			lastUserID := ""

			// Fetch all members with the role
			for {
				members, err := s.GuildMembers(i.GuildID, lastUserID, 1000)
				if err != nil {
					s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Content: "Failed to fetch guild members.",
						},
					})
					return
				}

				if len(members) == 0 {
					break
				}

				allMembers = append(allMembers, members...)
				lastUserID = members[len(members)-1].User.ID
			}

			var memberList []string
			for _, member := range allMembers {
				if contains(member.Roles, role.ID) {
					memberList = append(memberList, member.User.Username)
				}
			}

			// Send the list of members
			_, err = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
				Content: "Members with the role `@" + role.Name + "`:\n" + "```\n" + strings.Join(memberList, "\n") + "```",
				Flags:   discordgo.MessageFlagsEphemeral,
			})
		}
	}
}
