package main

import (
	"djs-zth-utilities/commands"
	"djs-zth-utilities/events"
	"djs-zth-utilities/posts"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"

	"github.com/bwmarrin/discordgo"
	"github.com/spf13/viper"
)

func onReady(s *discordgo.Session, event *discordgo.Ready) {
	log.Println("Bot is ready")

	guild, err := s.Guild(viper.GetString("guildID"))
	if err != nil {
		log.Println("Error getting guild:", err)
		return
	}

	// Set up cache for guild members
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		events.CacheGuildMembers(s, guild.ID)
	}()
	events.RegisterCommands(s)

	// Post the role selection embed
	err = posts.PostRoleSelectionEmbed(s)
	if err != nil {
		log.Printf("Error posting role selection embed: %v", err)
	}
	err = posts.PostKeySelectionEmbed(s)
	if err != nil {
		log.Printf("Error posting game selection embed: %v", err)
	}
	err = posts.PostValorSelectionEmbed(s)
	if err != nil {
		log.Printf("Error posting game selection embed: %v", err)
	}
	err = posts.PostPronounSelectionEmbed(s)
	if err != nil {
		log.Printf("Error posting pronoun selection embed: %v", err)
	}
	// Set up embed remover
	posts.EmbedRemover(s)
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

	intents := discordgo.IntentsGuildMembers | discordgo.IntentsGuildMessages | discordgo.IntentsGuildMessageReactions | discordgo.IntentsGuildMessageTyping | discordgo.IntentsGuilds | discordgo.IntentsGuildVoiceStates | discordgo.IntentsDirectMessages | discordgo.IntentsDirectMessageReactions | discordgo.IntentsDirectMessageTyping

	discord, err := discordgo.New("Bot " + viper.GetString("botToken"))
	if err != nil {
		log.Fatalf("error creating Discord session: %s", err)
	}
	discord.Identify.Intents = intents
	// Commands and events
	discord.AddHandler(onReady)
	discord.AddHandler(commands.InteractionCreate)
	discord.AddHandler(commands.AddRoleInteractionCreate)
	discord.AddHandler(commands.ListRole)
	discord.AddHandler(commands.RemoveRole)
	discord.AddHandler(commands.Suggestion)
	discord.AddHandler(events.RoleButtonInteractionCreate)
	discord.AddHandler(events.HandleReportMessageCommand)
	discord.AddHandler(events.OnDJsThreadCreate)
	discord.AddHandler(events.OnZthTicketCreate)
	discord.AddHandler(events.OnMemberJoin)
	discord.AddHandler(events.OnMemberLeave)
	discord.AddHandler(events.OnMemberUpdate)
	discord.AddHandler(events.OnMessageDelete)
	discord.AddHandler(events.OnMessageCreate)
	discord.AddHandler(events.OnMessageUpdate)
	discord.AddHandler(posts.HandleRoleSelection)
	discord.AddHandler(posts.HandleKeySelection)
	discord.AddHandler(posts.HandleValorSelection)
	discord.AddHandler(posts.HandlePronounSelection)

	discord.Open()
	defer discord.Close()

	log.Println("Bot is now running. Press CTRL+C to exit.")
	// Wait for a signal to exit
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	fmt.Println("Shutting down bot...")
}
