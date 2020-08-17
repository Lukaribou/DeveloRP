package main

import (
	"errors"
	"strings"

	"github.com/bwmarrin/discordgo"
)

// ConfigStruct : Contenu de config.json + compléments
type ConfigStruct struct {
	Token          string
	OwnerID        string
	GitHubLink     string
	InviteLink     string
	Prefix         string
	DbPassword     string
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
	Execute         func(Context)
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
	execute func(Context),
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
	ID         string
	userID     string
	money      int
	level      int
	createDate int64
	lastCode   int64
}

// ***************
