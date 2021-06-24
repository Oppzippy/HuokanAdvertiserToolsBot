package main

import (
	"bytes"

	"github.com/bwmarrin/discordgo"
)

func interactionCreateHandler(discord *discordgo.Session, event *discordgo.InteractionCreate) {
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
