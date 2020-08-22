package main

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

// ConfigStruct : Contenu de config.json + compléments
type ConfigStruct struct {
	Token          string   `json:"Token"`
	OwnerID        string   `json:"OwnerID"`
	GitHubLink     string   `json:"GitHubLink"`
	InviteLink     string   `json:"InviteLink"`
	Prefix         string   `json:"Prefix"`
	DbPassword     string   `json:"DbPassword"`
	SQLPort        []string `json:"SQLPort"`
	Version        string
	CommandHandler CommandHandler
}

// ***************

// Context : Ensemble de données pour les commandes
type Context struct {
	Guild   *discordgo.Guild
	Channel *discordgo.Channel
	User    *discordgo.User
	Member  *discordgo.Member
	Message *discordgo.MessageCreate
	Session *discordgo.Session
	DB      *DB
	Args    []string
}

// Reply : Envoie le texte donné dans le salon du message reçu
func (c *Context) Reply(msg string) (*discordgo.Message, error) {
	return c.Session.ChannelMessageSend(c.Channel.ID, msg)
}

// ReplyError : Envoie un embed avec l'erreur donnée
func (c *Context) ReplyError(msg string) (*discordgo.Message, error) {
	return c.Session.ChannelMessageSendEmbed(c.Channel.ID, &discordgo.MessageEmbed{
		Color:       0xFF0000,
		Description: XEMOJI + "**Erreur:**\n" + msg})
}

// ***************

// Command : Structure d'une commande
type Command struct {
	Name            string
	Category        string
	Aliases         []string
	Description     string
	GuildAdminsOnly bool
	OwnerOnly       bool
	Execute         func(*Context)
}

// ***************

// CommandHandler : ...
type CommandHandler struct {
	Commands []*Command
}

// Get : Renvoie la commande si elle existe, une erreur sinon
func (ch *CommandHandler) Get(name string) (*Command, error) {
	name = strings.ToLower(name)
	for _, c := range ch.Commands {
		if c.Name == name || ArrIncludes(c.Aliases, name) {
			return c, nil
		}
	}
	return &Command{}, errors.New("Command not found")
}

// AddCommand : Ajoute une commande au CommandHandler
func (ch *CommandHandler) AddCommand(name string,
	category string,
	aliases []string,
	description string,
	execute func(*Context),
	guildAdminsOnly bool,
	ownerOnly bool) {
	if aliases == nil {
		aliases = []string{}
	}
	ch.Commands = append(ch.Commands, &Command{name, category, aliases, description, guildAdminsOnly, ownerOnly, execute})
	Log("S", "Commande \"%s\" chargée.", name)
}

// ***************

// Player : Représente un joueur dans la BDD
type Player struct {
	ID          int
	userID      string
	money       uint
	xp          uint
	level       int
	createDate  int64
	lastCode    int64
	curLangName string
	skills      int

	db *DB
}

// HasSkill : Retourne true si le joueur a le skill
// Même fonctionnement que permissions Discord
func (pl *Player) HasSkill(code int) bool {
	return pl.skills&code != 0
}

// GetOwnedSkills : Renvoie la liste des skills que le joueur possède
func (pl *Player) GetOwnedSkills() []*Skill {
	skills := []*Skill{}
	for _, skill := range pl.db.GetSkills() {
		if pl.HasSkill(skill.gain) {
			skills = append(skills, skill)
		}
	}
	return skills
}

// GetTotalSkillsPoint : Renvoie le nombre de points communs au langage et au joueur
func (pl *Player) GetTotalSkillsPoint() int {
	total := 0
	for _, skill := range pl.GetOwnedSkills() {
		total += skill.gain
	}
	return total
}

// GetCurrentLanguage : Renvoie le langage actuel du joueur
func (pl *Player) GetCurrentLanguage() *Language {
	l, _ := pl.db.GetLanguage(pl.curLangName)
	return l
}

// ***************

// Language : Représente un langage dans la BDD
type Language struct {
	ID     int
	name   string
	level  int
	skills int
	cost   int
	imgURL string
	color  int

	db *DB
}

// HasSkill : Retourne true si le langage a le skill
// Même fonctionnement que permissions Discord
func (l *Language) HasSkill(code int) bool {
	return l.skills&code != 0
}

// SkillsCount : Renvoie le nombre de skills achetable dans le langage
func (l *Language) SkillsCount() int {
	count := 0
	s := strconv.FormatInt(int64(l.skills), 2)
	for i := 0; i < len(s); i++ {
		v, _ := strconv.Atoi(string(s[i]))
		if v%2 == 1 {
			count++
		}
	}
	return count
}

// GetSkills : Renvoie un tableau avec les skills du langage
func (l *Language) GetSkills() []*Skill {
	skills := []*Skill{}
	for _, skill := range l.db.GetSkills() {
		if l.HasSkill(skill.gain) {
			skills = append(skills, skill)
		}
	}
	return skills
}

// ***************

// Skill : Représente un skill dans la BDD
type Skill struct {
	ID   int
	cost int
	// Représente aussi le code du skill
	gain    int
	name    string
	special bool

	db *DB
}

// ***************

// https://github.com/CS-5/disgoreact/blob/master/disgoreact.go

type (
	// WatchContext : Les objets nécessaires pour watch un message
	WatchContext struct {
		Message  *discordgo.Message
		Session  *discordgo.Session
		TickRate time.Duration
		Context  *Context
		Data     interface{}
	}
	// WatchOption : Callback & expiration pour un émoji
	WatchOption struct {
		Emoji         string
		OnSuccess     func(user *discordgo.User, wCtx *WatchContext)
		OnError       func(err error, wCtx *WatchContext)
		LimitReaction int
		Expiration    time.Duration
		Filter        func(ctx *Context) bool
	}
)

// NewWatcher : Crée un nouveau Watcher. tickRate != 0
func NewWatcher(msg *discordgo.Message, ses *discordgo.Session, tickRate time.Duration, ctx *Context, data interface{}) *WatchContext {
	return &WatchContext{
		Message:  msg,
		Session:  ses,
		TickRate: tickRate,
		Context:  ctx,
		Data:     data,
	}
}

// Add : Ajoute un watcher au WatchContext.
// Les réactions sont dans un tableau d'Options
func (ctx *WatchContext) Add(options ...*WatchOption) {
	for _, v := range options {
		if err := ctx.Session.MessageReactionAdd(ctx.Message.ChannelID, ctx.Message.ID, v.Emoji); err != nil {
			Log("Err", "Impossible d'ajouter une réaction au message n°%s.", ctx.Message.ID)
		}
		go ctx.watcher(v)
	}
}

func (ctx *WatchContext) watcher(opt *WatchOption) {
	exp := time.After(opt.Expiration)
	tick := time.Tick(ctx.TickRate)
	expired := false

	for {
		select {
		case <-exp:
			expired = true
		case <-tick:
			if expired {
				ctx.Session.MessageReactionsRemoveAll(ctx.Message.ChannelID, ctx.Message.ID)
				return
			}
			if user, err := watchReactionPoll(ctx.Session, ctx.Message.ChannelID, ctx.Message.ID, opt); err != nil {
				opt.OnError(err, ctx)
				return
			} else if (discordgo.User{}) != *user && opt.Filter(ctx.Context) {
				opt.OnSuccess(user, ctx)
			}
		}
	}
}

func watchReactionPoll(s *discordgo.Session, chID, msgID string, opt *WatchOption) (*discordgo.User, error) {
	users, err := s.MessageReactions(chID, msgID, opt.Emoji, opt.LimitReaction, "", "")
	if err != nil {
		return &discordgo.User{}, err
	}

	if len(users) >= 1 {
		for _, u := range users {
			if u.ID == s.State.User.ID {
				continue
			}
			return u, nil
		}
	}
	return &discordgo.User{}, nil
}

// ***************
