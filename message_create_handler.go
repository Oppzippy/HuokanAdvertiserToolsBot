package main

import (
	"os"

	"github.com/bwmarrin/discordgo"
)

func newMessageCreateHandler(mc *DownloadButtonMessageCollection) func(discord *discordgo.Session, event *discordgo.MessageCreate) {
	return func(discord *discordgo.Session, event *discordgo.MessageCreate) {
		if event.GuildID != "" && event.Message.Content == "!huokanadvertisertools" {
			guild, err := discord.State.Guild(event.GuildID)
			if err != nil {
				errLogger.Printf("error fetching guild from state: %v", err)
				return
			}
			if guild.OwnerID != event.Author.ID {
				return
			}

			m, err := discord.ChannelMessageSendComplex(event.ChannelID, &discordgo.MessageSend{
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
			} else {
				mc.Add(DownloadButtonMessage{
					ChannelID: m.ChannelID,
					MessageID: m.ID,
				})
				f, err := os.Create("messages.json")
				if err != nil {
					errLogger.Printf("error creating messages file: %v", err)
				} else {
					err := mc.Write(f)
					if err != nil {
						errLogger.Printf("error writing messages file: %v", err)
					}
					f.Close()
				}
			}
		}
	}
}
