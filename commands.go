package main

import (
	"fmt"
	"strconv"
	"time"

	"github.com/bwmarrin/discordgo"
)

func pingCommand(ctx Context) {
	ctx.Reply("T'as cru que t'allais avoir la latence ?\nhttps://tenor.com/view/ha-cheh-take-that-gif-14055512")
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

func playerCreate(ctx Context) {
	if ctx.DB.PlayerExist(ctx.User.ID) {
		ctx.ReplyError("Vous possédez déjà un joueur.")
		return
	}

	id, _ := strconv.Atoi(ctx.User.ID)
	_, err := ctx.DB.sql.Exec("INSERT INTO users (userID, money, level, createDate, lastPay) VALUES (?, '0', 1, ?, nil)",
		id, strconv.Itoa(int(time.Now().Unix())))
	if err != nil {
		ctx.ReplyError("Une erreur SQL est survenue.")
		Log("Base de données | Erreur", "Erreur %s", err.Error())
		return
	}

	ctx.Reply(OKEMOJI + " **Vous êtes maintenant enregistré(e) dans ma base de données.**")
}

func displayPlayer(ctx Context) {
	var target *discordgo.User
	if len(ctx.Args) == 2 {
		if len(ctx.Message.Mentions) != 0 {
			target = ctx.Message.Mentions[0]
		} else {
			u, err := ctx.Session.User(ctx.Args[1])
			if err != nil {
				ctx.ReplyError("L'id donné ne correspond à aucun de mes utilisateurs.")
				return
			}
			target = u
		}
	} else {
		target = ctx.User
	}

	player, err := ctx.DB.GetPlayer(target.ID)
	if err != nil {
		ctx.ReplyError(err.Error())
		return
	}

	em := &discordgo.MessageEmbed{
		Color:  0x00FF00,
		Author: &discordgo.MessageEmbedAuthor{Name: "Informations sur " + target.Username, IconURL: target.AvatarURL("")},
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Bits:", Value: strconv.Itoa(player.money), Inline: true},
			{Name: "Niveau:", Value: strconv.Itoa(player.level), Inline: true},
			{Name: "Date de création:", Value: TimestampSecToDate(player.createDate), Inline: true}},
		Footer: &discordgo.MessageEmbedFooter{Text: "BDD ID: " + player.ID, IconURL: ctx.Session.State.User.AvatarURL("")},
	}

	if player.lastCode != 0 {
		em.Fields = append(em.Fields, &discordgo.MessageEmbedField{Name: "Dernier code:", Value: TimestampSecToDate(player.lastCode), Inline: true})
	}

	ctx.Session.ChannelMessageSendEmbed(ctx.Channel.ID, em)
}

func codeCommand(ctx Context) {
	now := time.Now()
	pl, err := ctx.DB.GetPlayer(ctx.User.ID)
	if err != nil {
		ctx.ReplyError("Vous ne possédez pas de joueur.")
		return
	}
	last := time.Unix(pl.lastCode, 0)

	if last.Add(6 * time.Hour).After(now) { // => Si ça fait moins de 6h
		ctx.ReplyError("Vous devez attendre *" + TimeFormatFr(last.Add(6*time.Hour)) + "* avant la prochaine session de code.")
		return
	}

	gain := 1 * pl.level

	if err = ctx.DB.sql.QueryRow("UPDATE users SET lastPay = ?, money = ? WHERE ID = ?",
		now.Unix(), pl.money+gain, pl.ID).Err(); err != nil {
		ctx.ReplyError("Une erreur SQL est survenue.")
		return
	}

	ctx.Reply(OKEMOJI + " **Votre session de code vous a fait gagner `" + strconv.Itoa(gain) + "` bits.**")
}
