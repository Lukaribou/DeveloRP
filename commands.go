package main

import "github.com/bwmarrin/discordgo"

func pingCommand(ctx Context) {
	ctx.Reply("Pong !")
}

func helpCommand(ctx Context) {
	var fields []*discordgo.MessageEmbedField
	for _, c := range Config.CommandHandler.Commands {
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:   c.Name,
			Value:  c.Description,
			Inline: true,
		})
	}

	em := &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			Name:    ctx.Session.State.User.Username + " - Page d'aide",
			IconURL: ctx.Session.State.User.AvatarURL(""),
			URL:     Config.GitHubLink},
		Fields: fields,
	}
	ctx.Session.ChannelMessageSendEmbed(ctx.Channel.ID, em)
}
