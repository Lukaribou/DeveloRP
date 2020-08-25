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
		"INSERT INTO users (userID, createDate) VALUES (?, ?)",
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

	pl, err := ctx.DB.GetPlayer(target.ID)
	if err != nil {
		ctx.ReplyError(err.Error())
		return
	}

	listSkills := []string{}
	for _, s := range pl.GetOwnedSkills() {
		listSkills = append(listSkills, s.name)
	}
	listSkillsStr := "`" + strings.Join(listSkills, "`, `") + "`"

	lang := pl.GetCurrentLanguage()

	em := &discordgo.MessageEmbed{
		Color:  lang.color,
		Author: &discordgo.MessageEmbedAuthor{Name: "Informations sur " + target.Username, IconURL: target.AvatarURL("")},
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Bits:", Value: strconv.FormatUint(uint64(pl.money), 10), Inline: true},
			{Name: "XP:", Value: strconv.FormatUint(uint64(pl.xp), 10) + " (Niveau:" + strconv.Itoa(pl.level) + ")", Inline: true},
			{Name: "Date de création:", Value: TimestampSecToDate(pl.createDate), Inline: true},
			{Name: "Langage actuel:", Value: lang.name + " (" + strconv.Itoa(lang.level) + ")", Inline: true},
			{Name: "Compétences:", Value: listSkillsStr + fmt.Sprintf(" (%d/%d)", len(listSkills), lang.SkillsCount()), Inline: true}},
		Footer:    &discordgo.MessageEmbedFooter{Text: "BDD ID: " + strconv.Itoa(pl.ID), IconURL: ctx.Session.State.User.AvatarURL("")},
		Thumbnail: &discordgo.MessageEmbedThumbnail{URL: lang.imgURL},
	}

	if pl.lastCode != 0 {
		em.Fields = append(em.Fields, &discordgo.MessageEmbedField{Name: "Dernier code:", Value: TimestampSecToDate(pl.lastCode), Inline: true})
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

	gain := pl.GetTotalSkillsPoint() * pl.GetCurrentLanguage().ID

	if _, err := ctx.DB.sql.Exec("UPDATE users SET lastCode = ?, money = ? WHERE ID = ?",
		time.Now().Unix(), pl.money+uint64(gain), pl.ID); err != nil {
		ctx.ReplyError("Une erreur SQL est survenue.")
		return
	}

	ctx.Reply(OKEMOJI + " **Votre session de code vous a fait gagner `" + strconv.Itoa(gain) + "` bits.**")
	if ne, e := pl.AddXP(100); ne && e == nil {
		ctx.Reply(TADAEMOJI + " **Vous venez de passer au niveau suivant !**")
	}
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
		}, OnError: func(err error, _ *WatchContext) {
			editMessageError(ctx, msg, err.Error())
		}, LimitReaction: 1, Expiration: 3e4, FilterUser: true,
	}, &WatchOption{
		Emoji: XEMOJI, OnSuccess: func(_ *discordgo.User, _ *WatchContext) {
			ctx.Session.ChannelMessageEdit(msg.ChannelID, msg.ID, XEMOJI+" **Requête SQL avortée.**")
		}, OnError: func(err error, _ *WatchContext) {
			editMessageError(ctx, msg, err.Error())
		}, LimitReaction: 1, Expiration: 3e4, FilterUser: true,
	})
}

func editMessageError(ctx *Context, msg *discordgo.Message, err string) {
	ctx.Session.ChannelMessageEdit(msg.ChannelID, msg.ID, XEMOJI+" **Une erreur est survenue :** "+err)
	Log("Err", "Erreur réaction collector : %s", err)
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

// BuyCommand : Commande pour acheter un skill / afficher le shop
func BuyCommand(ctx *Context) {
	pl, nPlErr := ctx.DB.GetPlayer(ctx.User.ID)
	if nPlErr != nil {
		ctx.ReplyError("Vous ne possédez pas de joueur.")
		return
	}
	if len(ctx.Args) == 1 {
		lang := pl.GetCurrentLanguage()
		fields := []*discordgo.MessageEmbedField{
			{Name: "🔰 Compétences", Value: "", Inline: true},
			{Name: "🔱 Capacités spéciales", Value: "Non disponible", Inline: true}}

		for _, skill := range ctx.DB.GetSkills() {
			nspe := ""
			t := fmt.Sprintf("`%d` (__%d__b) %s\n", skill.ID, skill.cost, skill.name)
			if pl.HasSkill(skill.gain) {
				nspe += OKEMOJI + t
			} else if !lang.HasSkill(skill.gain) {
				if skill.ID >= 15 {
					continue
				}
				nspe += XEMOJI + t
			} else {
				nspe += UNLOCKEDEMOJI + t
			}
			fields[0].Value += nspe
		}

		em := &discordgo.MessageEmbed{
			Author:      &discordgo.MessageEmbedAuthor{Name: "DeveloRP | Shop", IconURL: ctx.Session.State.User.AvatarURL(""), URL: Config.GitHubLink},
			Description: fmt.Sprintf("**Aide :** *%s Possédée / %s Achetable / %s Inachetable* `ID` (__Prix__ en bits) Nom", OKEMOJI, UNLOCKEDEMOJI, XEMOJI),
			Fields:      fields,
			Color:       lang.color,
			Thumbnail:   &discordgo.MessageEmbedThumbnail{URL: lang.imgURL},
			Footer:      &discordgo.MessageEmbedFooter{Text: fmt.Sprintf("Vous possédez %d bits | Langage : %s (%d)", pl.money, lang.name, lang.ID), IconURL: INFORMATIONSICON},
		}
		ctx.Session.ChannelMessageSendEmbed(ctx.Channel.ID, em)
	} else if len(ctx.Args) == 2 {
		asked, aErr := strconv.Atoi(ctx.Args[1])
		if aErr != nil {
			ctx.ReplyError("Le paramètre doit être l'ID de la compétence souhaitée. Faites `dv!shop` pour afficher les compétences et leurs ID.")
			return
		}
		skill, nSkillErr := ctx.DB.GetSkill(asked)
		if nSkillErr != nil {
			ctx.ReplyError(nSkillErr.Error() + ".")
			return
		} else if !skill.special && !pl.GetCurrentLanguage().HasSkill(asked) {
			ctx.ReplyError("Le langage `" + pl.curLangName + "` ne contient pas la compétence `" + skill.name + "`")
			return
		} else if pl.money < uint64(skill.cost) {
			ctx.ReplyError("La compétence `" + skill.name + "` coûte `" + strconv.Itoa(skill.cost) + "` bits. Vous n'en possédez que `" + strconv.FormatUint(pl.money, 10) + "`.")
			return
		} else if pl.HasSkill(skill.gain) {
			ctx.ReplyError("Vous possédez déjà cette compétence !")
			return
		}

		msg, mErr := ctx.Session.ChannelMessageSend(ctx.Channel.ID, "La compétence `"+skill.name+"` coûte `"+strconv.Itoa(skill.cost)+"`. Etes-vous sûr(e) de vouloir l'acheter ?")
		if mErr != nil {
			Log("Err", "Erreur BuyCommand 2 args : %s", mErr.Error())
			return
		}
		watcher := NewWatcher(msg, ctx.Session, 500, ctx, nil)
		watcher.Add(&WatchOption{
			Emoji: OKEMOJI, OnSuccess: func(_ *discordgo.User, _ *WatchContext) {
				err := pl.UpdateMoney(-skill.cost)
				if err != nil {
					ctx.ReplyError("Une erreur SQL est survenue.")
					Log("BDD Err", "Erreur SQL BuyCommand -> UpdateMoney : %s", err)
					return
				}
				err = pl.AddSkill(skill)
				if err != nil {
					ctx.ReplyError("Une erreur SQL est survenue. Vos bits vont vous être réstorés.")
					Log("BDD Err", "Erreur SQL BuyCommand -> AddSkill: %s", err)
					e := pl.UpdateMoney(skill.cost)
					if e != nil {
						Log("BDD Err", "Erreur BuyCommand -> Restoration argent : %s", e)
					}
					return
				}
				ctx.Session.ChannelMessageEdit(ctx.Channel.ID, msg.ID, OKEMOJI+" **Vous venez d'acquérir la compétence `"+skill.name+"` !**")
				if ne, e := pl.AddXP(200); ne && e == nil {
					ctx.Reply(TADAEMOJI + " **Vous venez de passer au niveau suivant !**")
				}
			}, OnError: func(err error, wCtx *WatchContext) {
				editMessageError(ctx, msg, err.Error())
			}, LimitReaction: 1, Expiration: 3e4, FilterUser: true,
		}, &WatchOption{
			Emoji: XEMOJI, OnSuccess: func(_ *discordgo.User, _ *WatchContext) {
				ctx.Session.ChannelMessageEdit(ctx.Channel.ID, msg.ID, XEMOJI+" **Vous avez annulé votre action.**")
			}, OnError: func(err error, _ *WatchContext) {
				editMessageError(ctx, msg, err.Error())
			}, LimitReaction: 1, Expiration: 3e4, FilterUser: true,
		})
	} else {
		ctx.Reply("Entrez la commande sans paramètre pour afficher le shop, ou la commande + l'id d'une compétence pour acheter celle-ci.")
	}
}

// DailyCommand : Commande qui rapporte un peu toutes les 24h
func DailyCommand(ctx *Context) {

}
