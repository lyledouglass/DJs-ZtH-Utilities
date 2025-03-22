package commands

import (
	"strings"

	"github.com/bwmarrin/discordgo"
)

func ListRole(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type == discordgo.InteractionApplicationCommand {
		switch i.ApplicationCommandData().Name {
		case "listrole":
			// Fetch the role from the interaction
			role := i.ApplicationCommandData().Options[0].RoleValue(s, i.GuildID)

			// Fetch all members with the role
			members, err := s.GuildMembers(i.GuildID, "", 1000)
			if err != nil {
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "Failed to fetch guild members.",
					},
				})
				return
			}

			var memberList []string
			for _, member := range members {
				if contains(member.Roles, role.ID) {
					memberList = append(memberList, member.User.Username)
				}
			}

			// Send the list of members
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Members with the role `@" + role.Name + "`:\n" + "```\n" + strings.Join(memberList, "\n") + "```",
				},
			})
		}
	}
}
