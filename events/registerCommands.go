package events

import (
	"log"
	"sync"

	"github.com/bwmarrin/discordgo"
	"github.com/spf13/viper"
)

func buildRaidTeamChoices() []*discordgo.ApplicationCommandOptionChoice {
	raidTeams, ok := viper.Get("raidTeams").([]interface{})
	if !ok {
		return nil
	}
	choices := make([]*discordgo.ApplicationCommandOptionChoice, 0, len(raidTeams))
	for _, t := range raidTeams {
		teamMap, ok := t.(map[string]interface{})
		if !ok {
			continue
		}
		name, _ := teamMap["name"].(string)
		value, _ := teamMap["value"].(string)
		if name != "" && value != "" {
			choices = append(choices, &discordgo.ApplicationCommandOptionChoice{
				Name:  name,
				Value: value,
			})
		}
	}
	return choices
}

func RegisterCommands(s *discordgo.Session) {
	teamChoices := buildRaidTeamChoices()
	gameChoices := []*discordgo.ApplicationCommandOptionChoice{
		{Name: "World of Warcraft", Value: "wow"},
		{Name: "Final Fantasy XIV", Value: "ffxiv"},
	}
	commands := []*discordgo.ApplicationCommand{
		{
			Name:        "ping",
			Description: "Respond with pong",
		},
		{
			Name:        "addrole",
			Description: "Add a role to a user",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionUser,
					Name:        "target",
					Description: "The user to add the role to",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionRole,
					Name:        "role",
					Description: "The role to add to the user",
					Required:    true,
				},
			},
		},
		{
			Name:        "removerole",
			Description: "Remove a role from a user",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionUser,
					Name:        "target",
					Description: "The user to remove the role from",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionRole,
					Name:        "role",
					Description: "The role to remove from the user",
					Required:    true,
				},
			},
		},
		{
			Name:        "listrole",
			Description: "List all members with a role",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionRole,
					Name:        "role",
					Description: "The role to list members for",
					Required:    true,
				},
			},
		},
		{
			Name: "Report Message",
			Type: discordgo.MessageApplicationCommand,
		},
		{
			Name:        "create-raid-team-info",
			Description: "Create a raid team info post in the raid teams channel",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "team",
					Description: "The raid team",
					Required:    true,
					Choices:     teamChoices,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "game",
					Description: "The game this team plays",
					Required:    true,
					Choices:     gameChoices,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "schedule",
					Description: "Raid days and times (e.g. Tue/Thu 8-11pm EST)",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "current-prog",
					Description: "Current progression (e.g. 8/8 Mythic)",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "recruitment-contact",
					Description: "Discord username(s) to contact for recruitment",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "blurb",
					Description: "Team description, goals, expectations, etc.",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "currently-recruiting",
					Description: "Who the team is recruiting (e.g. DPS, Holy Paladin, or Not Recruiting)",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "app-link",
					Description: "Link to the team application (optional)",
					Required:    false,
				},
			},
		},
		{
			Name:        "update-raid-team-info",
			Description: "Update an existing raid team info post",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "team",
					Description: "The raid team to update",
					Required:    true,
					Choices:     teamChoices,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "game",
					Description: "The game this team plays",
					Required:    false,
					Choices:     gameChoices,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "schedule",
					Description: "Raid days and times",
					Required:    false,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "current-prog",
					Description: "Current progression",
					Required:    false,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "recruitment-contact",
					Description: "Discord username(s) to contact for recruitment",
					Required:    false,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "blurb",
					Description: "Team description, goals, expectations, etc.",
					Required:    false,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "currently-recruiting",
					Description: "Who the team is recruiting (e.g. DPS, Holy Paladin, or Not Recruiting)",
					Required:    false,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "app-link",
					Description: "Link to the team application",
					Required:    false,
				},
			},
		},
		{
			Name:        "suggestion",
			Description: "Submit a suggestion for the server",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "suggestion",
					Description: "The suggestion to submit",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "team",
					Description: "The team to send the suggestion to",
					Required:    true,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{
							Name:  "Rocket",
							Value: "rocket_leadership",
						},
						{
							Name:  "Gravity",
							Value: "gravity_leadership",
						},
						{
							Name:  "Phoenix",
							Value: "phoenix_leadership",
						},
						{
							Name:  "Ramrod",
							Value: "ramrod_leadership",
						},
						{
							Name:  "Death Jesters",
							Value: "death_jesters_leadership",
						},
						{
							Name:  "Integrity",
							Value: "integrity_leadership",
						},
						{
							Name:  "Eclipse",
							Value: "eclipse_leadership",
						},
						{
							Name:  "Last Call",
							Value: "last_call_leadership",
						},
					},
				},
			},
		},
	}
	var wg sync.WaitGroup
	for _, command := range commands {
		wg.Add(1)
		go func(cmd *discordgo.ApplicationCommand) {
			defer wg.Done()
			_, err := s.ApplicationCommandCreate(s.State.User.ID, "", cmd)
			log.Println("Registering command:", cmd.Name)
			if err != nil {
				log.Fatal("Cannot create command: ", err)
			}
		}(command)
	}
}
