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
	// Test Command
	discord.AddHandler(onReady)
	discord.AddHandler(commands.InteractionCreate)
	discord.AddHandler(commands.AddRoleInteractionCreate)

	discord.Open()
	defer discord.Close()

	fmt.Println("Bot is now running. Press CTRL+C to exit.")
	// Wait for a signal to exit
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	fmt.Println("Shutting down bot...")

}
