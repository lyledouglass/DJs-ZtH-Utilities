package events

import (
	"log"
	"sync"

	"github.com/bwmarrin/discordgo"
)

func RegisterCommands(s *discordgo.Session) {
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
