package posts

import (
	"fmt"
	"log"
	"sort"
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

// sortRoleOptions sorts role options by grouping similar roles together
func sortRoleOptions(options []discordgo.SelectMenuOption) {
	sort.Slice(options, func(i, j int) bool {
		labelI := options[i].Label
		labelJ := options[j].Label

		// Extract full category (e.g., "Normal Raid", "Heroic Raid", "Low Key", etc.)
		getFullCategory := func(label string) (string, int) {
			// Define category priorities
			categoryPriorities := map[string]int{
				"Normal Raid":   1,
				"Heroic Raid":   2,
				"Mythic Raid":   3,
				"Low Key":       4,
				"Mid Key":       5,
				"High Key":      6,
				"Hero Key":      7,
				"Normal Delves": 8,
				"Hero Delves":   9,
				"Retail BGs":    10,
				"Retail Arena":  11,
				"RGB":           12,
				"Valor Farm":    13,
				"Collectors":    14,
			}

			// Check for each category in priority order
			for category, priority := range categoryPriorities {
				if strings.Contains(label, category) {
					return category, priority
				}
			}

			// Handle special cases for older naming
			if strings.Contains(label, "BGs") && !strings.Contains(label, "Retail") {
				return "Retail BGs", 10
			}
			if strings.Contains(label, "Arena") && !strings.Contains(label, "Retail") && !strings.Contains(label, "RGB") {
				return "Retail Arena", 11
			}

			return "Unknown", 999
		}

		// Get role type priority
		getRoleTypePriority := func(label string) int {
			if strings.Contains(label, "Tank") {
				return 1
			}
			if strings.Contains(label, "Heals") {
				return 2
			}
			if strings.Contains(label, "DPS") {
				return 3
			}
			return 4 // For roles without specific type
		}

		// Get categories for both labels
		catI, priI := getFullCategory(labelI)
		catJ, priJ := getFullCategory(labelJ)

		// First sort by category priority
		if priI != priJ {
			return priI < priJ
		}

		// If same category, sort by role type (Tank, Heals, DPS)
		if catI == catJ {
			roleTypeI := getRoleTypePriority(labelI)
			roleTypeJ := getRoleTypePriority(labelJ)

			if roleTypeI != roleTypeJ {
				return roleTypeI < roleTypeJ
			}
		}

		// Finally, sort alphabetically as fallback
		return labelI < labelJ
	})
}

// PostRoleSelectionEmbed posts an embed with a dropdown for raid and PvP role selection
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

	// Filter for raid and PvP roles only
	var options []discordgo.SelectMenuOption
	raidPvpKeywords := []string{"Raid", "BGs", "Arena", "RGB"}
	for roleID, roleName := range openRoles {
		for _, keyword := range raidPvpKeywords {
			if strings.Contains(roleName, keyword) {
				options = append(options, discordgo.SelectMenuOption{
					Label: roleName,
					Value: roleID,
				})
				break
			}
		}
	}

	// Sort the options
	sortRoleOptions(options)

	embed := &discordgo.MessageEmbed{
		Title:       "LFG Role Selection",
		Description: "Select roles for raiding and PvP content. Once selected, you'll be able to be pinged in the <#" + viper.GetString("lfgChannelId") + "> channel for relevant group content.",
		Color:       0x00ff00,
	}

	components := []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.SelectMenu{
					CustomID:    "role_select",
					Placeholder: "Choose roles...",
					Options:     options,
					MinValues:   &[]int{0}[0],
					MaxValues:   len(options),
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

// PostKeySelectionEmbed posts an embed with a dropdown for key role selection
func PostKeySelectionEmbed(s *discordgo.Session) error {
	channelID := viper.GetString("roleSelectionChannelId")
	openRoles := viper.GetStringMapString("openRoles")

	if channelID == "" {
		return fmt.Errorf("roleSelectionChannelId not configured")
	}

	if len(openRoles) == 0 {
		return fmt.Errorf("openRoles not configured")
	}

	// Check if embed already exists
	exists, err := checkEmbedExists(s, channelID, "Mythic+ Key Selection")
	if err != nil {
		log.Printf("Error checking if key selection embed exists: %v", err)
	} else if exists {
		log.Println("Key selection embed already exists, skipping post")
		return nil
	}

	// Filter for key roles only
	var options []discordgo.SelectMenuOption
	keyKeywords := []string{"Key", "Delves"}
	for roleID, roleName := range openRoles {
		for _, keyword := range keyKeywords {
			if strings.Contains(roleName, keyword) {
				options = append(options, discordgo.SelectMenuOption{
					Label: roleName,
					Value: roleID,
				})
				break
			}
		}
	}

	// Sort the options
	sortRoleOptions(options)

	embed := &discordgo.MessageEmbed{
		Title:       "Mythic+ Key Selection",
		Description: "Select roles for Mythic+ dungeons and delves. Choose the difficulty levels you're comfortable with to get pinged for relevant groups in <#" + viper.GetString("lfgChannelId") + ">.",
		Color:       0xff6600,
	}

	components := []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.SelectMenu{
					CustomID:    "key_select",
					Placeholder: "Choose key roles...",
					Options:     options,
					MinValues:   &[]int{0}[0],
					MaxValues:   len(options),
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

// PostValorSelectionEmbed posts an embed with a dropdown for valor and collection roles
func PostValorSelectionEmbed(s *discordgo.Session) error {
	channelID := viper.GetString("roleSelectionChannelId")
	openRoles := viper.GetStringMapString("openRoles")

	if channelID == "" {
		return fmt.Errorf("roleSelectionChannelId not configured")
	}

	if len(openRoles) == 0 {
		return fmt.Errorf("openRoles not configured")
	}

	// Check if embed already exists
	exists, err := checkEmbedExists(s, channelID, "Other Activities Selection")
	if err != nil {
		log.Printf("Error checking if other activities selection embed exists: %v", err)
	} else if exists {
		log.Println("Other activities selection embed already exists, skipping post")
		return nil
	}

	// Filter for valor and collection roles
	var options []discordgo.SelectMenuOption
	otherKeywords := []string{"Valor", "Collectors"}
	for roleID, roleName := range openRoles {
		for _, keyword := range otherKeywords {
			if strings.Contains(roleName, keyword) {
				options = append(options, discordgo.SelectMenuOption{
					Label: roleName,
					Value: roleID,
				})
				break
			}
		}
	}

	// Sort the options
	sortRoleOptions(options)

	embed := &discordgo.MessageEmbed{
		Title:       "Other Activities Selection",
		Description: "Select roles for other activities like valor farming and collecting. Get pinged for these activities in <#" + viper.GetString("lfgChannelId") + ">.",
		Color:       0x9900ff,
	}

	components := []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.SelectMenu{
					CustomID:    "valor_select",
					Placeholder: "Choose activity roles...",
					Options:     options,
					MinValues:   &[]int{0}[0],
					MaxValues:   len(options),
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

	// Process multiple role selections
	var addedRoles, removedRoles []string
	selectedRoleIDs := data.Values

	// Get all raid/PvP roles for comparison
	openRoles := viper.GetStringMapString("openRoles")
	raidPvpKeywords := []string{"Raid", "BGs", "Arena", "RGB"}
	var allRaidPvpRoles []string
	for roleID, roleName := range openRoles {
		for _, keyword := range raidPvpKeywords {
			if strings.Contains(roleName, keyword) {
				allRaidPvpRoles = append(allRaidPvpRoles, roleID)
				break
			}
		}
	}

	// Check which roles user currently has
	userRoles := make(map[string]bool)
	for _, role := range member.Roles {
		userRoles[role] = true
	}

	// Process each raid/PvP role
	for _, roleID := range allRaidPvpRoles {
		hasRole := userRoles[roleID]
		shouldHaveRole := false

		// Check if this role is selected
		for _, selectedRole := range selectedRoleIDs {
			if selectedRole == roleID {
				shouldHaveRole = true
				break
			}
		}

		if hasRole && !shouldHaveRole {
			// Remove role
			err = s.GuildMemberRoleRemove(guildID, userID, roleID)
			if err != nil {
				log.Printf("Failed to remove role %s from user %s: %v", roleID, userID, err)
			} else {
				removedRoles = append(removedRoles, openRoles[roleID])
				log.Printf("Successfully removed role %s from user %s", roleID, userID)
			}
		} else if !hasRole && shouldHaveRole {
			// Add role
			err = s.GuildMemberRoleAdd(guildID, userID, roleID)
			if err != nil {
				log.Printf("Failed to add role %s to user %s: %v", roleID, userID, err)
			} else {
				addedRoles = append(addedRoles, openRoles[roleID])
				log.Printf("Successfully added role %s to user %s", roleID, userID)
			}
		}
	}

	// Recreate the select menu components to keep it interactive
	var options []discordgo.SelectMenuOption
	for rID, roleName := range openRoles {
		for _, keyword := range raidPvpKeywords {
			if strings.Contains(roleName, keyword) {
				options = append(options, discordgo.SelectMenuOption{
					Label: roleName,
					Value: rID,
				})
				break
			}
		}
	}

	// Sort the options
	sortRoleOptions(options)

	components := []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.SelectMenu{
					CustomID:    "role_select",
					Placeholder: "Choose roles...",
					Options:     options,
					MinValues:   &[]int{0}[0],
					MaxValues:   len(options),
				},
			},
		},
	}

	// Update the original message to reset the select menu
	embed := &discordgo.MessageEmbed{
		Title:       "Raid and PvP Role Selection",
		Description: "Select roles for raiding and PvP content. Once selected, you'll be able to be pinged in the <#" + viper.GetString("lfgChannelId") + "> channel for relevant group content.",
		Color:       0x00ff00,
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Embeds:     []*discordgo.MessageEmbed{embed},
			Components: components,
		},
	})

	// Create response message
	var responseText string
	if len(addedRoles) > 0 || len(removedRoles) > 0 {
		if len(addedRoles) > 0 {
			responseText += "Added: " + strings.Join(addedRoles, ", ")
		}
		if len(removedRoles) > 0 {
			if responseText != "" {
				responseText += "\n"
			}
			responseText += "Removed: " + strings.Join(removedRoles, ", ")
		}
	} else {
		responseText = "No role changes made."
	}

	// Send ephemeral follow-up message to the user
	s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
		Content: responseText,
		Flags:   discordgo.MessageFlagsEphemeral,
	})
}

// HandleKeySelection handles the key selection interaction
func HandleKeySelection(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionMessageComponent {
		return
	}

	data := i.MessageComponentData()
	if data.CustomID != "key_select" {
		return
	}

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

	// Process multiple role selections
	var addedRoles, removedRoles []string
	selectedRoleIDs := data.Values

	// Get all key roles for comparison
	openRoles := viper.GetStringMapString("openRoles")
	keyKeywords := []string{"Key", "Delves"}
	var allKeyRoles []string
	for roleID, roleName := range openRoles {
		for _, keyword := range keyKeywords {
			if strings.Contains(roleName, keyword) {
				allKeyRoles = append(allKeyRoles, roleID)
				break
			}
		}
	}

	// Check which roles user currently has
	userRoles := make(map[string]bool)
	for _, role := range member.Roles {
		userRoles[role] = true
	}

	// Process each key role
	for _, roleID := range allKeyRoles {
		hasRole := userRoles[roleID]
		shouldHaveRole := false

		// Check if this role is selected
		for _, selectedRole := range selectedRoleIDs {
			if selectedRole == roleID {
				shouldHaveRole = true
				break
			}
		}

		if hasRole && !shouldHaveRole {
			// Remove role
			err = s.GuildMemberRoleRemove(guildID, userID, roleID)
			if err != nil {
				log.Printf("Failed to remove key role %s from user %s: %v", roleID, userID, err)
			} else {
				removedRoles = append(removedRoles, openRoles[roleID])
				log.Printf("Successfully removed key role %s from user %s", roleID, userID)
			}
		} else if !hasRole && shouldHaveRole {
			// Add role
			err = s.GuildMemberRoleAdd(guildID, userID, roleID)
			if err != nil {
				log.Printf("Failed to add key role %s to user %s: %v", roleID, userID, err)
			} else {
				addedRoles = append(addedRoles, openRoles[roleID])
				log.Printf("Successfully added key role %s to user %s", roleID, userID)
			}
		}
	}

	// Recreate the select menu components to keep it interactive
	var options []discordgo.SelectMenuOption
	for rID, roleName := range openRoles {
		for _, keyword := range keyKeywords {
			if strings.Contains(roleName, keyword) {
				options = append(options, discordgo.SelectMenuOption{
					Label: roleName,
					Value: rID,
				})
				break
			}
		}
	}

	// Sort the options
	sortRoleOptions(options)

	components := []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.SelectMenu{
					CustomID:    "key_select",
					Placeholder: "Choose key roles...",
					Options:     options,
					MinValues:   &[]int{0}[0],
					MaxValues:   len(options),
				},
			},
		},
	}

	// Update the original message to reset the select menu
	embed := &discordgo.MessageEmbed{
		Title:       "Mythic+ Key Selection",
		Description: "Select roles for Mythic+ dungeons and delves. Choose the difficulty levels you're comfortable with to get pinged for relevant groups in <#" + viper.GetString("lfgChannelId") + ">.",
		Color:       0xff6600,
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Embeds:     []*discordgo.MessageEmbed{embed},
			Components: components,
		},
	})

	// Create response message
	var responseText string
	if len(addedRoles) > 0 || len(removedRoles) > 0 {
		if len(addedRoles) > 0 {
			responseText += "Added: " + strings.Join(addedRoles, ", ")
		}
		if len(removedRoles) > 0 {
			if responseText != "" {
				responseText += "\n"
			}
			responseText += "Removed: " + strings.Join(removedRoles, ", ")
		}
	} else {
		responseText = "No role changes made."
	}

	// Send ephemeral follow-up message to the user
	s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
		Content: responseText,
		Flags:   discordgo.MessageFlagsEphemeral,
	})
}

// HandleValorSelection handles the valor/collection selection interaction
func HandleValorSelection(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionMessageComponent {
		return
	}

	data := i.MessageComponentData()
	if data.CustomID != "valor_select" {
		return
	}

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

	// Process multiple role selections
	var addedRoles, removedRoles []string
	selectedRoleIDs := data.Values

	// Get all valor/collection roles for comparison
	openRoles := viper.GetStringMapString("openRoles")
	otherKeywords := []string{"Valor", "Collectors"}
	var allOtherRoles []string
	for roleID, roleName := range openRoles {
		for _, keyword := range otherKeywords {
			if strings.Contains(roleName, keyword) {
				allOtherRoles = append(allOtherRoles, roleID)
				break
			}
		}
	}

	// Check which roles user currently has
	userRoles := make(map[string]bool)
	for _, role := range member.Roles {
		userRoles[role] = true
	}

	// Process each valor/collection role
	for _, roleID := range allOtherRoles {
		hasRole := userRoles[roleID]
		shouldHaveRole := false

		// Check if this role is selected
		for _, selectedRole := range selectedRoleIDs {
			if selectedRole == roleID {
				shouldHaveRole = true
				break
			}
		}

		if hasRole && !shouldHaveRole {
			// Remove role
			err = s.GuildMemberRoleRemove(guildID, userID, roleID)
			if err != nil {
				log.Printf("Failed to remove activity role %s from user %s: %v", roleID, userID, err)
			} else {
				removedRoles = append(removedRoles, openRoles[roleID])
				log.Printf("Successfully removed activity role %s from user %s", roleID, userID)
			}
		} else if !hasRole && shouldHaveRole {
			// Add role
			err = s.GuildMemberRoleAdd(guildID, userID, roleID)
			if err != nil {
				log.Printf("Failed to add activity role %s to user %s: %v", roleID, userID, err)
			} else {
				addedRoles = append(addedRoles, openRoles[roleID])
				log.Printf("Successfully added activity role %s to user %s", roleID, userID)
			}
		}
	}

	// Recreate the select menu components to keep it interactive
	var options []discordgo.SelectMenuOption
	for rID, roleName := range openRoles {
		for _, keyword := range otherKeywords {
			if strings.Contains(roleName, keyword) {
				options = append(options, discordgo.SelectMenuOption{
					Label: roleName,
					Value: rID,
				})
				break
			}
		}
	}

	// Sort the options
	sortRoleOptions(options)

	components := []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.SelectMenu{
					CustomID:    "valor_select",
					Placeholder: "Choose activity roles...",
					Options:     options,
					MinValues:   &[]int{0}[0],
					MaxValues:   len(options),
				},
			},
		},
	}

	// Update the original message to reset the select menu
	embed := &discordgo.MessageEmbed{
		Title:       "Other Activities Selection",
		Description: "Select roles for other activities like valor farming and collecting. Get pinged for these activities in <#" + viper.GetString("lfgChannelId") + ">.",
		Color:       0x9900ff,
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Embeds:     []*discordgo.MessageEmbed{embed},
			Components: components,
		},
	})

	// Create response message
	var responseText string
	if len(addedRoles) > 0 || len(removedRoles) > 0 {
		if len(addedRoles) > 0 {
			responseText += "Added: " + strings.Join(addedRoles, ", ")
		}
		if len(removedRoles) > 0 {
			if responseText != "" {
				responseText += "\n"
			}
			responseText += "Removed: " + strings.Join(removedRoles, ", ")
		}
	} else {
		responseText = "No role changes made."
	}

	// Send ephemeral follow-up message to the user
	s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
		Content: responseText,
		Flags:   discordgo.MessageFlagsEphemeral,
	})
}
