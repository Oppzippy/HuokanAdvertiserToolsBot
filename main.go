package main

import (
	"archive/zip"
	"bytes"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		errLogger.Printf("Error loading .env file: %v", err)
	}
	discord, err := discordgo.New("Bot " + os.Getenv("DISCORD_TOKEN"))
	if err != nil {
		errLogger.Fatalf("Error creating discord connection: %v", err)
	}
	defer discord.Close()
	discord.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentsGuilds

	discord.AddHandler(MessageCreateHandler)
	discord.AddHandler(InteractionCreateHandler)

	discord.Open()

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM)
	<-sc
	logger.Println("Stopping bot")
}

func MessageCreateHandler(discord *discordgo.Session, event *discordgo.MessageCreate) {
	if event.GuildID != "" && event.Message.Content == "!huokanadvertisertools" {
		guild, err := discord.State.Guild(event.GuildID)
		if err != nil {
			errLogger.Printf("error fetching guild from state: %v", err)
			return
		}
		if guild.OwnerID != event.Author.ID {
			return
		}

		_, err = discord.ChannelMessageSendComplex(event.ChannelID, &discordgo.MessageSend{
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
			errLogger.Printf("Error sending message with download button: %v", err)
		}
	}
}

func InteractionCreateHandler(discord *discordgo.Session, event *discordgo.InteractionCreate) {
	if event.Type == discordgo.InteractionMessageComponent &&
		event.MessageComponentData().CustomID == "download_addon" &&
		event.Member != nil {

		customScript := NewCustomScript()
		customScript.SetDiscordTag(event.Member.User.String())
		addon, err := getCustomizedAddon(customScript)
		if err != nil {
			errLogger.Printf("failed to create custom zip: %v", err)
			return
		}
		fileName := "HuokanAdvertiserTools.zip"
		if addon.Version != "" {
			fileName = "HuokanAdvertiserTools-" + addon.Version + ".zip"
		}
		_, err = sendDM(discord, event.Member.User.ID, &discordgo.MessageSend{
			Files: []*discordgo.File{
				{
					Name:        fileName,
					ContentType: "application/zip",
					Reader:      bytes.NewReader(addon.Content),
				},
			},
		})

		response := "Check your DMs!"
		if err != nil {
			response = "Failed to send DM. Please make sure you are allowing DMs from server members."
		} else {
			logger.Printf("Sent addon to %s", event.Member.User.String())
		}

		err = discord.InteractionRespond(event.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: response,
				Flags:   64, // Ephemeral
			},
		})
		if err != nil {
			errLogger.Printf("Error sending interaction response: %v", err)
		}
	}
}

func sendDM(discord *discordgo.Session, userID string, message *discordgo.MessageSend) (*discordgo.Message, error) {
	dmChannel, err := discord.UserChannelCreate(userID)
	if err != nil {
		return nil, err
	}
	sentMessage, err := discord.ChannelMessageSendComplex(dmChannel.ID, message)
	return sentMessage, err
}

func getCustomizedAddon(customScript *CustomScript) (*PackagedAddon, error) {
	unmodifiedZip, err := zip.OpenReader("HuokanAdvertiserTools.zip")
	if err != nil {
		return nil, fmt.Errorf("error opening base addon zip: %v", err)
	}
	defer unmodifiedZip.Close()
	addon, err := Package(&unmodifiedZip.Reader, customScript)
	if err != nil {
		return nil, fmt.Errorf("error packaging addon: %v", err)
	}
	return addon, nil
}
