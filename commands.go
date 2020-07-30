package main

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

func pingCommand(ctx Context) {
	ctx.Reply("Pong !")
}

func helpCommand(ctx Context) {
	categories := make(map[string][]*Command, 0)
	for _, c := range Config.CommandHandler.Commands {
		categories[c.Category] = append(categories[c.Category], c)
	}

	owner, oErr := ctx.Session.User(Config.OwnerID)

	for categ, commands := range categories {
		var fields []*discordgo.MessageEmbedField
		for _, c := range commands {
			fields = append(fields, &discordgo.MessageEmbedField{
				Name:  c.Name,
				Value: fmt.Sprintf("%s\n*BotOwnerOnly : %s / GuildAdminsOnly :* %s", c.Description, GetEmojiOkOrX(c.OwnerOnly), GetEmojiOkOrX(c.GuildAdminsOnly)),
			})
		}

		em := &discordgo.MessageEmbed{
			Author: &discordgo.MessageEmbedAuthor{
				Name:    ctx.Session.State.User.Username + " - Page d'aide\nCatégorie : " + categ,
				IconURL: ctx.Session.State.User.AvatarURL(""),
				URL:     Config.GitHubLink},
			Color:  RandomColor(),
			Fields: fields,
		}

		if oErr == nil {
			em.Footer = &discordgo.MessageEmbedFooter{
				Text:    fmt.Sprintf("Prefix: %s | ©%s (%s)", Config.Prefix, owner.Username+"#"+owner.Discriminator, Config.OwnerID),
				IconURL: owner.AvatarURL(""),
			}
		}

		dm, _ := ctx.Session.UserChannelCreate(ctx.User.ID)
		ctx.Session.ChannelMessageSendEmbed(dm.ID, em)
	}

	ctx.Session.ChannelMessageSendEmbed(ctx.Channel.ID, &discordgo.MessageEmbed{
		Title: OKEMOJI + " L'aide vous a été envoyée en MP",
		Color: 0xFFFFFF,
	})
}
