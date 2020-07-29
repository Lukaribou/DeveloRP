package main

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

func pingCommand(ctx Context) {
	ctx.Reply("Pong !")
}

func helpCommand(ctx Context) {
	var fields []*discordgo.MessageEmbedField
	for _, c := range Config.CommandHandler.Commands {
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:   c.Name,
			Value:  c.Description,
			Inline: true})
	}

	em := &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			Name:    ctx.Session.State.User.Username + " - Page d'aide",
			IconURL: ctx.Session.State.User.AvatarURL(""),
			URL:     Config.GitHubLink},
		Fields: fields}

	if owner, err := ctx.Session.User(Config.OwnerID); err == nil {
		em.Footer = &discordgo.MessageEmbedFooter{
			Text:    fmt.Sprintf("Prefix: %s | ©%s (%s)", Config.Prefix, owner.Username+owner.Discriminator, Config.OwnerID),
			IconURL: owner.AvatarURL("")}
	}

	ctx.Session.ChannelMessageSendEmbed(ctx.Channel.ID, em)
}
