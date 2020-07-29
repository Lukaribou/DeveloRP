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
	CommandHandler CommandHandler
}

// ***************

// Context : Ensemble de données pour les commandes
type Context struct {
	Guild   *discordgo.Guild
	Channel *discordgo.Channel
	Member  *discordgo.Member
	Message *discordgo.MessageCreate
	Session *discordgo.Session
	Args    []string
}

// Reply : Envoie le texte donné dans le salon du message reçu
func (c *Context) Reply(msg string) (*discordgo.Message, error) {
	return c.Session.ChannelMessageSend(c.Channel.ID, msg)
}

// ***************

// Command : Structure d'une commande
type Command struct {
	Name            string
	Aliases         []string
	Description     string
	GuildAdminsOnly bool
	OwnerOnly       bool
	Execute         func(Context)
}

// CommandHandler : ...
type CommandHandler struct {
	Commands []Command
}

// Get : Renvoie la commande si elle existe, une erreur sinon
func (ch *CommandHandler) Get(name string) (Command, error) {
	name = strings.ToLower(name)
	for _, c := range ch.Commands {
		if c.Name == name || ArrIncludes(c.Aliases, name) {
			return c, nil
		}
	}
	return Command{}, errors.New("Command not found")
}

// AddCommand : Ajoute une commande au CommandHandler
func (ch *CommandHandler) AddCommand(name string,
	aliases []string,
	description string,
	execute func(Context),
	guildAdminsOnly bool,
	ownerOnly bool) {
	if aliases == nil {
		aliases = make([]string, 0)
	}
	ch.Commands = append(ch.Commands, Command{name, aliases, description, guildAdminsOnly, ownerOnly, execute})
	Log("Système", "Commande \"%s\" chargée.\n", name)
}
