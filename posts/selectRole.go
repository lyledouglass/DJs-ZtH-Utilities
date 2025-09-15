package posts

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/spf13/viper"
)

// checkEmbedExists checks if an embed with the given title already exists in the channel
func checkEmbedExists(s *discordgo.Session, channelID, embedTitle string) (bool, error) {
	messages, err := s.ChannelMessages(channelID, 50, "", "", "")
	if err != nil {
		return false, err
	}

	for _, message := range messages {
		if len(message.Embeds) > 0 && message.Embeds[0].Title == embedTitle {
			return true, nil
		}
	}
	return false, nil
}

// PostRoleSelectionEmbed posts an embed with a dropdown for role selection
func PostRoleSelectionEmbed(s *discordgo.Session) error {
	channelID := viper.GetString("roleSelectionChannelId")
	openRoles := viper.GetStringMapString("openRoles")

	if channelID == "" {
		return fmt.Errorf("roleSelectionChannelId not configured")
	}

	if len(openRoles) == 0 {
		return fmt.Errorf("openRoles not configured")
	}

	// Check if embed already exists
	exists, err := checkEmbedExists(s, channelID, "LFG Role Selection")
	if err != nil {
		log.Printf("Error checking if role selection embed exists: %v", err)
	} else if exists {
		log.Println("Role selection embed already exists, skipping post")
		return nil
	}

	// Create select menu options
	var options []discordgo.SelectMenuOption
	for roleID, roleName := range openRoles {
		options = append(options, discordgo.SelectMenuOption{
			Label: roleName,
			Value: roleID,
		})
	}

	embed := &discordgo.MessageEmbed{
		Title: "LFG Role Selection",
		Description: "We have several roles available to community members to utilize for forming groups for many types of group content. Whether it is mythic plus keystones, raiding, PvP, or other in game content, we have a role for that (or let us know if we don't!) This is a great way for you to meet other community members and new friends! \n \n" +
			"Please click on the dropdowns below and select the roles you would like to be pinged for in the <#" + viper.GetString("lfgChannelId") + "> channel. Once selected, click outside of the dropdown box and you will see a confirmation message at the bottom of the channel. We suggest you uncheck the roles or suppress notifications for all roles for the channel when you don't want to be pinged. Please state the intention/goal of a group when pinging someone.\n \n" + "Use /lfg to get started with finding a group for keys, or start a new post in the unknown channel with a good description and you will able to ping roles from your post.",
		Color: 0x00ff00,
	}

	components := []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.SelectMenu{
					CustomID:    "role_select",
					Placeholder: "Choose a role...",
					Options:     options,
				},
			},
		},
	}

	_, err = s.ChannelMessageSendComplex(channelID, &discordgo.MessageSend{
		Embeds:     []*discordgo.MessageEmbed{embed},
		Components: components,
	})

	return err
}

// HandleRoleSelection handles the role selection interaction
func HandleRoleSelection(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionMessageComponent {
		return
	}

	data := i.MessageComponentData()
	if data.CustomID != "role_select" {
		return
	}

	if len(data.Values) == 0 {
		return
	}

	roleID := data.Values[0]
	userID := i.Member.User.ID
	guildID := i.GuildID

	// Helper function to get member with retry
	getMemberWithRetry := func() (*discordgo.Member, error) {
		for attempts := 0; attempts < 3; attempts++ {
			if attempts > 0 {
				time.Sleep(time.Millisecond * 200) // Small delay for cache to update
			}
			member, err := s.GuildMember(guildID, userID)
			if err != nil {
				return nil, err
			}
			return member, nil
		}
		return nil, fmt.Errorf("failed to get member after retries")
	}

	// Fetch current member data to get up-to-date roles
	member, err := getMemberWithRetry()
	if err != nil {
		log.Printf("Error fetching member data for user %s: %v", userID, err)
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Failed to fetch your current roles.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	// Check if user has the role
	hasRole := false
	for _, role := range member.Roles {
		if role == roleID {
			hasRole = true
			break
		}
	}

	log.Printf("User %s role check: hasRole=%t, roleID=%s, memberRoles=%v", userID, hasRole, roleID, member.Roles)

	var responseText string

	if hasRole {
		// Remove role
		err = s.GuildMemberRoleRemove(guildID, userID, roleID)
		if err != nil {
			responseText = "Failed to remove role."
			log.Printf("Failed to remove role %s from user %s: %v", roleID, userID, err)
		} else {
			responseText = "Role removed successfully!"
			log.Printf("Successfully removed role %s from user %s", roleID, userID)
		}
	} else {
		// Add role
		err = s.GuildMemberRoleAdd(guildID, userID, roleID)
		if err != nil {
			responseText = "Failed to add role."
			log.Printf("Failed to add role %s to user %s: %v", roleID, userID, err)
		} else {
			responseText = "Role added successfully!"
			log.Printf("Successfully added role %s to user %s", roleID, userID)
		}
	}

	// Recreate the select menu components to keep it interactive
	openRoles := viper.GetStringMapString("openRoles")
	var options []discordgo.SelectMenuOption
	for rID, roleName := range openRoles {
		options = append(options, discordgo.SelectMenuOption{
			Label: roleName,
			Value: rID,
		})
	}

	components := []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.SelectMenu{
					CustomID:    "role_select",
					Placeholder: "Choose a role...",
					Options:     options,
				},
			},
		},
	}

	// Update the original message to reset the select menu
	embed := &discordgo.MessageEmbed{
		Title: "LFG Role Selection",
		Description: "We have several roles available to community members to utilize for forming groups for many types of group content. Whether it is mythic plus keystones, raiding, PvP, or other in game content, we have a role for that (or let us know if we don't!) This is a great way for you to meet other community members and new friends! \n \n" +
			"Please click on the dropdowns below and select the roles you would like to be pinged for in the <#" + viper.GetString("lfgChannelId") + "> channel. Once selected, click outside of the dropdown box and you will see a confirmation message at the bottom of the channel. We suggest you uncheck the roles or suppress notifications for all roles for the channel when you don't want to be pinged. Please state the intention/goal of a group when pinging someone.\n \n" + "Use /lfg to get started with finding a group for keys, or start a new post in the unknown channel with a good description and you will able to ping roles from your post.",
		Color: 0x00ff00,
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Embeds:     []*discordgo.MessageEmbed{embed},
			Components: components,
		},
	})

	// Send ephemeral follow-up message to the user
	s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
		Content: responseText,
		Flags:   discordgo.MessageFlagsEphemeral,
	})
}

// PostGameSelectionEmbed posts an embed with buttons for game selection
func PostGameSelectionEmbed(s *discordgo.Session) error {
	channelID := viper.GetString("gameSelectionChannelId")
	gameRoles := viper.GetStringMapString("gameRoles")

	if channelID == "" {
		return fmt.Errorf("gameSelectionChannelId not configured")
	}

	if len(gameRoles) == 0 {
		return fmt.Errorf("gameRoles not configured")
	}

	// Check if embed already exists
	exists, err := checkEmbedExists(s, channelID, "Game Selection")
	if err != nil {
		log.Printf("Error checking if game selection embed exists: %v", err)
	} else if exists {
		log.Println("Game selection embed already exists, skipping post")
		return nil
	}

	embed := &discordgo.MessageEmbed{
		Title:       "Game Selection",
		Description: "Select which games you're interested in to get notified about relevant content and events. You can toggle these roles on and off as needed.",
		Color:       0x0099ff,
	}

	// Create buttons for each game
	var buttons []discordgo.MessageComponent
	for roleID, gameName := range gameRoles {
		buttons = append(buttons, discordgo.Button{
			Label:    gameName,
			Style:    discordgo.PrimaryButton,
			CustomID: "game_" + roleID,
		})
	}

	components := []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: buttons,
		},
	}

	_, err = s.ChannelMessageSendComplex(channelID, &discordgo.MessageSend{
		Embeds:     []*discordgo.MessageEmbed{embed},
		Components: components,
	})

	return err
}

// HandleGameSelection handles the game selection button interactions
func HandleGameSelection(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionMessageComponent {
		return
	}

	data := i.MessageComponentData()
	if !strings.HasPrefix(data.CustomID, "game_") {
		return
	}

	// Extract role ID from custom ID
	roleID := strings.TrimPrefix(data.CustomID, "game_")
	userID := i.Member.User.ID
	guildID := i.GuildID

	// Helper function to get member with retry
	getMemberWithRetry := func() (*discordgo.Member, error) {
		for attempts := 0; attempts < 3; attempts++ {
			if attempts > 0 {
				time.Sleep(time.Millisecond * 200)
			}
			member, err := s.GuildMember(guildID, userID)
			if err != nil {
				return nil, err
			}
			return member, nil
		}
		return nil, fmt.Errorf("failed to get member after retries")
	}

	// Fetch current member data to get up-to-date roles
	member, err := getMemberWithRetry()
	if err != nil {
		log.Printf("Error fetching member data for user %s: %v", userID, err)
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Failed to fetch your current roles.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	// Check if user has the role
	hasRole := false
	for _, role := range member.Roles {
		if role == roleID {
			hasRole = true
			break
		}
	}

	log.Printf("User %s game role check: hasRole=%t, roleID=%s, memberRoles=%v", userID, hasRole, roleID, member.Roles)

	var responseText string

	if hasRole {
		// Remove role
		err = s.GuildMemberRoleRemove(guildID, userID, roleID)
		if err != nil {
			responseText = "Failed to remove game role."
			log.Printf("Failed to remove game role %s from user %s: %v", roleID, userID, err)
		} else {
			responseText = "Game role removed successfully!"
			log.Printf("Successfully removed game role %s from user %s", roleID, userID)
		}
	} else {
		// Add role
		err = s.GuildMemberRoleAdd(guildID, userID, roleID)
		if err != nil {
			responseText = "Failed to add game role."
			log.Printf("Failed to add game role %s to user %s: %v", roleID, userID, err)
		} else {
			responseText = "Game role added successfully!"
			log.Printf("Successfully added game role %s to user %s", roleID, userID)
		}
	}

	// Recreate the button components to keep them interactive
	gameRoles := viper.GetStringMapString("gameRoles")
	var buttons []discordgo.MessageComponent
	for rID, gameName := range gameRoles {
		buttons = append(buttons, discordgo.Button{
			Label:    gameName,
			Style:    discordgo.PrimaryButton,
			CustomID: "game_" + rID,
		})
	}

	components := []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: buttons,
		},
	}

	// Update the original message to reset the buttons
	embed := &discordgo.MessageEmbed{
		Title:       "Game Selection",
		Description: "Select which games you're interested in to get notified about relevant content and events. You can toggle these roles on and off as needed.",
		Color:       0x0099ff,
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Embeds:     []*discordgo.MessageEmbed{embed},
			Components: components,
		},
	})

	// Send ephemeral follow-up message to the user
	s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
		Content: responseText,
		Flags:   discordgo.MessageFlagsEphemeral,
	})
}
