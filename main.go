package main

import (
	"archive/zip"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

func main() {
	logger.Println("Starting bot")
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

	messageCollection, err := loadMessages()
	if err != nil {
		errLogger.Fatalf("error loading messages: %v", err)
	}

	done, err := watchMessages(discord, messageCollection)

	if err != nil {
		errLogger.Printf("error watching addon zip for changes: %v", err)
	} else {
		defer close(done)
	}

	discord.AddHandler(newMessageCreateHandler(messageCollection))
	discord.AddHandler(interactionCreateHandler)

	err = discord.Open()
	if err != nil {
		errLogger.Fatalf("error connecting to discord: %v", err)
	}

	err = updateMessages(discord, messageCollection)
	if err != nil {
		errLogger.Printf("error updating messages: %v", err)
	}

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM)
	<-sc
	logger.Println("Stopping bot")
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

func loadMessages() (*DownloadButtonMessageCollection, error) {
	messageCollection := NewDownloadButtonMessageCollection()
	f, err := os.Open("messages.json")
	if err != nil {
		pathError, ok := err.(*os.PathError)
		if !ok || pathError.Err != syscall.ENOENT {
			return nil, err
		}
	}
	defer f.Close()
	messageCollection.Read(f)
	return messageCollection, nil
}

func watchMessages(discord *discordgo.Session, mc *DownloadButtonMessageCollection) (done chan struct{}, err error) {
	done, err = watch(".", "HuokanAdvertiserTools.zip", func() {
		err := updateMessages(discord, mc)
		if err != nil {
			errLogger.Printf("error updating messages after version change: %v", err)
		}
	})

	return done, err
}

func updateMessages(discord *discordgo.Session, mc *DownloadButtonMessageCollection) error {
	addon, err := getCustomizedAddon(NewCustomScript())
	if err != nil {
		return fmt.Errorf("error creating custom addon zip: %v", err)
	}
	for _, message := range mc.Messages() {
		edit := discordgo.NewMessageEdit(message.ChannelID, message.MessageID)
		edit.Components = []discordgo.MessageComponent{
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.Button{
						Label:    "Download Addon",
						Style:    discordgo.PrimaryButton,
						CustomID: "download_addon",
					},
				},
			},
		}
		content := "Huokan Advertiser Tools " + addon.Version
		edit.Content = &content
		_, err := discord.ChannelMessageEditComplex(edit)
		if err != nil {
			restErr, ok := err.(*discordgo.RESTError)
			if ok && restErr.Message.Code == discordgo.ErrCodeUnknownMessage {
				mc.Remove(message)
				f, err := os.Create("messages.json")
				if err != nil {
					return fmt.Errorf("error creating messages file: %v", err)
				}
				err = mc.Write(f)
				if err != nil {
					return fmt.Errorf("error writing messages: %v", err)
				}
				f.Close()
			} else {
				return fmt.Errorf("error updating version: %v", err)
			}
		}
	}
	return nil
}
