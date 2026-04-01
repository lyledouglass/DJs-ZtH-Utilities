package commands

import (
	"fmt"
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/spf13/viper"
)

func CreateRaidTeamInfo(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionApplicationCommand {
		return
	}
	if i.ApplicationCommandData().Name != "create-raid-team-info" {
		return
	}

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags: discordgo.MessageFlagsEphemeral,
		},
	})
	if err != nil {
		return
	}

	optionMap := buildOptionMap(i.ApplicationCommandData().Options)

	teamValue := optionMap["team"].StringValue()
	teamName := getRaidTeamDisplayName(teamValue)
	game := optionMap["game"].StringValue()
	schedule := optionMap["schedule"].StringValue()
	currentProg := optionMap["current-prog"].StringValue()
	recruitmentContact := optionMap["recruitment-contact"].StringValue()
	blurb := optionMap["blurb"].StringValue()
	currentlyRecruiting := optionMap["currently-recruiting"].StringValue()

	appLink := ""
	if opt, ok := optionMap["app-link"]; ok {
		appLink = opt.StringValue()
	}

	channelId := viper.GetString("raidTeamsChannelId")
	embed := buildRaidTeamEmbed(teamName, teamValue, game, schedule, currentProg, recruitmentContact, appLink, blurb, currentlyRecruiting)

	_, err = s.ChannelMessageSendEmbed(channelId, embed)
	if err != nil {
		log.Printf("Error sending raid team info embed: %v", err)
		s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Content: "Error creating raid team info. Please try again later.",
			Flags:   discordgo.MessageFlagsEphemeral,
		})
		return
	}

	s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
		Content: fmt.Sprintf("Raid team info for **%s** has been created!", teamName),
		Flags:   discordgo.MessageFlagsEphemeral,
	})
}

func UpdateRaidTeamInfo(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionApplicationCommand {
		return
	}
	if i.ApplicationCommandData().Name != "update-raid-team-info" {
		return
	}

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags: discordgo.MessageFlagsEphemeral,
		},
	})
	if err != nil {
		return
	}

	optionMap := buildOptionMap(i.ApplicationCommandData().Options)
	teamValue := optionMap["team"].StringValue()
	teamName := getRaidTeamDisplayName(teamValue)
	channelId := viper.GetString("raidTeamsChannelId")

	msg, existingEmbed, err := findRaidTeamEmbed(s, channelId, teamValue)
	if err != nil {
		log.Printf("Error searching for raid team embed: %v", err)
		s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Content: "Error searching for existing raid team info. Please try again later.",
			Flags:   discordgo.MessageFlagsEphemeral,
		})
		return
	}
	if msg == nil {
		s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Content: fmt.Sprintf("No existing raid team info found for **%s**. Use `/create-raid-team-info` first.", teamName),
			Flags:   discordgo.MessageFlagsEphemeral,
		})
		return
	}

	// Extract current field values from the existing embed
	currentFields := make(map[string]string)
	for _, f := range existingEmbed.Fields {
		currentFields[f.Name] = f.Value
	}

	game := parseGameFromFooter(existingEmbed.Footer.Text)
	schedule := currentFields["Schedule"]
	currentProg := currentFields["Current Progression"]
	recruitmentContact := currentFields["Recruitment Contact"]
	appLink := currentFields["Application Link"]
	currentlyRecruiting := currentFields["Currently Recruiting"]
	blurb := existingEmbed.Description

	// Override with any provided options
	if opt, ok := optionMap["game"]; ok {
		game = opt.StringValue()
	}
	if opt, ok := optionMap["schedule"]; ok {
		schedule = opt.StringValue()
	}
	if opt, ok := optionMap["current-prog"]; ok {
		currentProg = opt.StringValue()
	}
	if opt, ok := optionMap["recruitment-contact"]; ok {
		recruitmentContact = opt.StringValue()
	}
	if opt, ok := optionMap["app-link"]; ok {
		appLink = opt.StringValue()
	}
	if opt, ok := optionMap["currently-recruiting"]; ok {
		currentlyRecruiting = opt.StringValue()
	}
	if opt, ok := optionMap["blurb"]; ok {
		blurb = opt.StringValue()
	}

	updatedEmbed := buildRaidTeamEmbed(teamName, teamValue, game, schedule, currentProg, recruitmentContact, appLink, blurb, currentlyRecruiting)

	_, err = s.ChannelMessageEditComplex(&discordgo.MessageEdit{
		ID:      msg.ID,
		Channel: channelId,
		Embeds:  &[]*discordgo.MessageEmbed{updatedEmbed},
	})
	if err != nil {
		log.Printf("Error updating raid team info embed: %v", err)
		s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Content: "Error updating raid team info. Please try again later.",
			Flags:   discordgo.MessageFlagsEphemeral,
		})
		return
	}

	s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
		Content: fmt.Sprintf("Raid team info for **%s** has been updated!", teamName),
		Flags:   discordgo.MessageFlagsEphemeral,
	})
}

func buildRaidTeamEmbed(teamName, teamValue, game, schedule, currentProg, recruitmentContact, appLink, blurb, currentlyRecruiting string) *discordgo.MessageEmbed {
	gameKey := strings.ToLower(game)
	color := 0xFF8C00 // WoW orange
	gameLabel := "World of Warcraft"
	if gameKey == "ffxiv" {
		color = 0x6E3FF3 // FFXIV purple
		gameLabel = "Final Fantasy XIV"
	}

	thumbnailURL := viper.GetString("raidTeamGameThumbnails." + gameKey)

	fields := []*discordgo.MessageEmbedField{
		{Name: "Game", Value: gameLabel, Inline: true},
		{Name: "Schedule", Value: schedule, Inline: true},
		{Name: "Current Progression", Value: currentProg, Inline: false},
		{Name: "Recruitment Contact", Value: recruitmentContact, Inline: true},
		{Name: "Currently Recruiting", Value: currentlyRecruiting, Inline: true},
	}

	if appLink != "" {
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:   "Application Link",
			Value:  appLink,
			Inline: false,
		})
	}

	embed := &discordgo.MessageEmbed{
		Title:       teamName,
		Description: blurb,
		Color:       color,
		Fields:      fields,
		Footer: &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("team:%s | game:%s", teamValue, gameKey),
		},
	}

	if thumbnailURL != "" {
		embed.Thumbnail = &discordgo.MessageEmbedThumbnail{URL: thumbnailURL}
	}

	return embed
}

// findRaidTeamEmbed searches the channel for an existing embed posted by the bot for the given team.
func findRaidTeamEmbed(s *discordgo.Session, channelId, teamValue string) (*discordgo.Message, *discordgo.MessageEmbed, error) {
	footerPrefix := "team:" + teamValue + " |"
	var beforeID string

	for batch := 0; batch < 10; batch++ {
		msgs, err := s.ChannelMessages(channelId, 100, beforeID, "", "")
		if err != nil {
			return nil, nil, err
		}
		if len(msgs) == 0 {
			break
		}

		for _, msg := range msgs {
			if msg.Author.ID != s.State.User.ID {
				continue
			}
			for _, embed := range msg.Embeds {
				if embed.Footer != nil && strings.HasPrefix(embed.Footer.Text, footerPrefix) {
					return msg, embed, nil
				}
			}
		}

		beforeID = msgs[len(msgs)-1].ID
		if len(msgs) < 100 {
			break
		}
	}

	return nil, nil, nil
}

// parseGameFromFooter extracts the game value from the embed footer text.
// Footer format: "team:{teamValue} | game:{gameValue}"
func parseGameFromFooter(footerText string) string {
	parts := strings.Split(footerText, " | game:")
	if len(parts) == 2 {
		return parts[1]
	}
	return "wow"
}

// getRaidTeamDisplayName looks up the display name for a team value from config.
func getRaidTeamDisplayName(teamValue string) string {
	raidTeams, ok := viper.Get("raidTeams").([]interface{})
	if !ok {
		return teamValue
	}
	for _, t := range raidTeams {
		teamMap, ok := t.(map[string]interface{})
		if !ok {
			continue
		}
		if teamMap["value"] == teamValue {
			if name, ok := teamMap["name"].(string); ok {
				return name
			}
		}
	}
	return teamValue
}

func buildOptionMap(options []*discordgo.ApplicationCommandInteractionDataOption) map[string]*discordgo.ApplicationCommandInteractionDataOption {
	m := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
	for _, opt := range options {
		m[opt.Name] = opt
	}
	return m
}
