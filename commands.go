package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

// PingCommand : Voir si le bot répond
func PingCommand(ctx *Context) {
	ctx.Reply("T'as cru que t'allais avoir la latence ?\nhttps://tenor.com/view/ha-cheh-take-that-gif-14055512")
}

// HelpCommand : Afficher l'aide
func HelpCommand(ctx *Context) {
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

// PlayerCreate : Créer son joueur dans la BDD
func PlayerCreate(ctx *Context) {
	if ctx.DB.PlayerExist(ctx.User.ID) {
		ctx.ReplyError("Vous possédez déjà un joueur.")
		return
	}

	id, _ := strconv.Atoi(ctx.User.ID)
	_, err := ctx.DB.sql.Exec(
		"INSERT INTO users (userID, money, level, createDate, lastCode, curLangName, skills) VALUES (?, '0', 1, ?, 0, 'Python', 1)",
		id, strconv.Itoa(int(time.Now().Unix())))
	if err != nil {
		ctx.ReplyError("Une erreur SQL est survenue.")
		Log("BDD Err", "Erreur %s", err.Error())
		return
	}

	ctx.Reply(OKEMOJI + " **Vous êtes maintenant enregistré(e) dans ma base de données.**")
}

// DisplayPlayer : Afficher des informations sur son joueur
func DisplayPlayer(ctx *Context) {
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
		Footer: &discordgo.MessageEmbedFooter{Text: "BDD ID: " + strconv.Itoa(player.ID), IconURL: ctx.Session.State.User.AvatarURL("")},
	}

	if player.lastCode != 0 {
		em.Fields = append(em.Fields, &discordgo.MessageEmbedField{Name: "Dernier code:", Value: TimestampSecToDate(player.lastCode), Inline: true})
	}

	ctx.Session.ChannelMessageSendEmbed(ctx.Channel.ID, em)
}

// CodeCommand : Miner des bits (recevoir de l'argent)
func CodeCommand(ctx *Context) {
	pl, err := ctx.DB.GetPlayer(ctx.User.ID)
	if err != nil {
		ctx.ReplyError("Vous ne possédez pas de joueur.")
		return
	}

	if pl.lastCode != 0 {
		last := time.Unix(pl.lastCode, 0)
		if last.Add(time.Hour).After(time.Now()) { // => Si ça fait moins de 6h
			ctx.ReplyError("Vous devez attendre *" + TimeFormatFr(last.Add(time.Hour)) + "* avant la prochaine session de code.")
			return
		}
	}

	gain := pl.GetTotalSkillsPoint() * pl.level

	if _, err := ctx.DB.sql.Exec("UPDATE users SET lastCode = ?, money = ? WHERE ID = ?",
		time.Now().Unix(), pl.money+gain, pl.ID); err != nil {
		ctx.ReplyError("Une erreur SQL est survenue.")
		return
	}

	ctx.Reply(OKEMOJI + " **Votre session de code vous a fait gagner `" + strconv.Itoa(gain) + "` bits.**")
}

// ExecSQLCommand : Exécuter une commande SQL
func ExecSQLCommand(ctx *Context) {
	request := strings.Join(ctx.Args[1:], " ")
	msg, err := ctx.Session.ChannelMessageSend(ctx.Channel.ID, "Etes-vous sûr de vouloir exécuter la requête suivante ? ```"+request+"```")
	if err != nil {
		Log("Err", err.Error())
		return
	}
	watcher := NewWatcher(msg, ctx.Session, 500, ctx, nil)
	watcher.Add(&WatchOption{
		Emoji: OKEMOJI, OnSuccess: func(_ *discordgo.User, _ *WatchContext) {
			r, err := ctx.DB.sql.Exec(request)
			if err != nil {
				ctx.ReplyError("Une erreur est survenue. J'envoie les détails en MP.")
				if c, e := ctx.Session.UserChannelCreate(ctx.User.ID); e == nil {
					ctx.Session.ChannelMessageSend(c.ID, "**Erreur dans `execSQLCommand()`:**\n```"+err.Error()+"```")
				}
				Log("BDD Err", err.Error())
				return
			}

			msgText := fmt.Sprintf("%s **BDD mise à jour avec succès.**\nCommande executée: `%s`", OKEMOJI, request)
			if ra, e := r.RowsAffected(); e == nil {
				msgText += fmt.Sprintf("\nColonnes affectées: `%d`", ra)
			}
			ctx.Session.ChannelMessageEdit(msg.ChannelID, msg.ID, msgText)
		}, OnError: func(err error, wCtx *WatchContext) {
			editMessageError(ctx, msg, "**Une erreur est survenue.**")
			Log("Err", "Erreur réaction collector : %s", err.Error())
		}, LimitReaction: 1, Expiration: 3e4, Filter: func(_ *Context) bool { return true },
	}, &WatchOption{
		Emoji: XEMOJI, OnSuccess: func(_ *discordgo.User, _ *WatchContext) {
			ctx.Session.ChannelMessageEdit(msg.ChannelID, msg.ID, XEMOJI+" **Requête SQL avortée.**")
		}, OnError: func(err error, wCtx *WatchContext) {
			editMessageError(ctx, msg, "**Une erreur est survenue.**")
			Log("Err", "Erreur réaction collector : %s", err.Error())
		}, LimitReaction: 1, Expiration: 3e4, Filter: func(_ *Context) bool { return true },
	})
}

func editMessageError(ctx *Context, msg *discordgo.Message, err string) {
	ctx.Session.ChannelMessageEdit(msg.ChannelID, msg.ID, XEMOJI+" "+err)
}

// ShutdownCommand : Eteindre le bot et la BDD
func ShutdownCommand(ctx *Context) {
	ctx.Session.ChannelMessageSend(ctx.Channel.ID, OKEMOJI+" **Extinction du bot en cours...**")
	fmt.Println()
	Log("Sys", "Arrêt système demandé.")

	errSC := ctx.Session.Close()
	if errSC != nil {
		Log("Sys Err", "Erreur pendant la fermeture de la session: %s.", errSC.Error())
	}
	errSQLC := ctx.DB.sql.Close()
	if errSQLC != nil {
		Log("BDD Err", "Erreur pendant la fermeture de la base de données: %s.", errSQLC.Error())
	}

	if errSC == nil && errSQLC == nil {
		Log("Sys S", "Bot arrêté sans problème.")
	}
	os.Exit(0)
}
