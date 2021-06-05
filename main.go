package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Printf("Error loading .env file: %v", err)
	}
	discord, err := discordgo.New("Bot " + os.Getenv("DISCORD_TOKEN"))
	if err != nil {
		log.Fatalf("Error creating discord connection: %v", err)
	}
	defer discord.Close()

	discord.AddHandler(func(_ *discordgo.Session, event *discordgo.MessageCreate) {
		if event.GuildID != "" && event.Message.Content == "!huokanadvertisertools" {
			_, err := discord.ChannelMessageSendComplex(event.ChannelID, &discordgo.MessageSend{
				Content: "Huokan Advertiser Tools",
				Components: []discordgo.MessageComponent{
					discordgo.ActionsRow{
						Components: []discordgo.MessageComponent{
							discordgo.Button{
								Label:    "Download Addon",
								Style:    discordgo.PrimaryButton,
								CustomID: "download_addon",
							},
						},
					},
				},
			})
			if err != nil {
				log.Fatalf("Error sending message with download button: %v", err)
			}
		}
	})

	discord.AddHandler(func(_ *discordgo.Session, event *discordgo.InteractionCreate) {
		if event.Type == discordgo.InteractionMessageComponent && event.MessageComponentData().CustomID == "download_addon" {
			err := discord.InteractionRespond(event.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "downloading.",
					Flags:   64, // Ephemeral
				},
			})
			if err != nil {
				log.Fatalf("Error sending interaction response: %v", err)
			}
		}
	})

	discord.Open()

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM)
	<-sc
	log.Println("Stopping bot")
}
