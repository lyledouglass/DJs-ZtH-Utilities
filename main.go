package main

import (
	"djs-zth-utilities/commands"
	"djs-zth-utilities/events"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/bwmarrin/discordgo"
	"github.com/spf13/viper"
)

func onReady(s *discordgo.Session, event *discordgo.Ready) {
	log.Println("Bot is ready")
	events.RegisterCommands(s)

	// Set the bot's status to total members with the community role
	guild, err := s.Guild(viper.GetString("guildID"))
	if err != nil {
		log.Println("Error getting guild:", err)
		return
	}
	members, err := s.GuildMembers(guild.ID, "", 1000)
	if err != nil {
		log.Println("Error getting guild members:", err)
		return
	}
	roleID := viper.GetString("communityMemberRole")
	count := 0
	for _, member := range members {
		for _, role := range member.Roles {
			if role == roleID {
				count++
				break
			}
		}
	}
	totalMembers := count
	status := fmt.Sprintf("Community Memebers: %d", totalMembers)
	err = s.UpdateCustomStatus(status)
	if err != nil {
		log.Println("Error setting status:", err)
	}
}

func main() {
	// Load configuration
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalf("Error reading config file: %s", err)
	}

	discord, err := discordgo.New("Bot " + viper.GetString("botToken"))
	if err != nil {
		log.Fatalf("error creating Discord session: %s", err)
	}
	// Commands and events
	discord.AddHandler(onReady)
	discord.AddHandler(commands.InteractionCreate)
	discord.AddHandler(commands.AddRoleInteractionCreate)
	discord.AddHandler(commands.ListRole)
	discord.AddHandler(commands.RemoveRole)
	discord.AddHandler(events.RoleButtonInteractionCreate)
	discord.AddHandler(events.HandleReportMessageCommand)
	discord.AddHandler(events.OnDJsThreadCreate)

	discord.Open()
	defer discord.Close()

	fmt.Println("Bot is now running. Press CTRL+C to exit.")
	// Wait for a signal to exit
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	fmt.Println("Shutting down bot...")

}
