package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
)

var (
	// Config : La config du bot
	Config ConfigStruct

	commandsCooldown []string
)

func main() {
	readConfig()
	dg, err := discordgo.New("Bot " + Config.Token)
	if err != nil {
		log.Fatal("Erreur lors de la création du client")
	}

	dg.AddHandler(ready)
	dg.AddHandler(messageCreate)
	dg.AddHandler(guildCreate)
	dg.AddHandler(guildDelete)

	Config.DB = NewDB()
	defer Config.DB.sql.Close()

	rand.Seed(time.Now().UnixNano()) // Initialiser le rand
	Log("Sys S", "Générateur du paquet \"rand\" initialisé.")

	err = dg.Open()
	if err != nil {
		log.Fatal("Erreur lors de l'ouverture de la connection")
	}
	defer dg.Close()

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
}

func ready(s *discordgo.Session, e *discordgo.Ready) {
	s.UpdateListeningStatus(Config.Prefix + " | " + Config.Version)
	Log("Sys S", "Bot en ligne sur %d serveur(s) sous le nom de %s.\n", len(e.Guilds), e.User.Username)
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.Bot {
		return
	} else if c, _ := s.Channel(m.ChannelID); c.Type == discordgo.ChannelTypeDM || c.Type == discordgo.ChannelTypeGroupDM {
		s.ChannelMessageSend(c.ID, "**Je ne prends pas en compte les commandes effectuées en MP.\nSi vous le cherchez, mon prefix est `"+Config.Prefix+"`**")
		return
	} else if strings.TrimSpace(m.Content) == "<@!"+s.State.User.ID+">" {
		s.ChannelMessageSend(c.ID, "**Mon prefix est `"+Config.Prefix+"`**")
		return
	} else if !strings.HasPrefix(strings.ToLower(m.Content), Config.Prefix) {
		return
	}

	if isInCommandsCooldown(m.Author.ID) {
		s.ChannelMessageSend(m.ChannelID, XEMOJI+" **Le bot possède un cooldown de 1s.**")
		return
	}
	commandsCooldown = append(commandsCooldown, m.Author.ID)
	time.AfterFunc(time.Second, func() {
		commandsCooldown = ArrayRemove(commandsCooldown, ArrayFind(commandsCooldown, m.Author.ID))
	})

	m.Content = m.Content[len(Config.Prefix):]

	args := strings.Split(m.Content, " ")
	cmd, err := Config.CommandHandler.Get(strings.ToLower(args[0]))
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, XEMOJI+" **La commande `"+args[0]+"` m'est inconnue.**")
		return
	}

	if cmd.OwnerOnly && Config.OwnerID != m.Author.ID {
		s.ChannelMessageSend(m.ChannelID, XEMOJI+" **Cette commande est limitée au propriétaire du bot.**")
		return
	} else if p, _ := MemberHasPermission(s, m.GuildID, m.Author.ID, discordgo.PermissionAdministrator); cmd.GuildAdminsOnly && !p {
		s.ChannelMessageSend(m.ChannelID, XEMOJI+" **Cette commande est réservée aux administrateurs du serveur.**")
		return
	}
	g, err := s.Guild(m.GuildID)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, XEMOJI+" **Une erreur est survenue.**")
		return
	}

	if c, err := s.Channel(m.ChannelID); err == nil {
		go cmd.Execute(&Context{g, c, m.Author, m.Member, m, s, Config.DB, args})
	}
}

func guildCreate(s *discordgo.Session, g *discordgo.GuildCreate) {
	owner, _ := s.User(g.OwnerID)
	txt := fmt.Sprintf("Serveur rejoint : %s (Owner : %s (ID: %s)) avec %d membres.", g.Name, owner.Username+"#"+owner.Discriminator, g.OwnerID, g.MemberCount)
	Log("Système | Avertissement", txt)
	LogFile("Système | Avertissement", "", txt)
}

func guildDelete(s *discordgo.Session, g *discordgo.GuildDelete) {
	owner, _ := s.User(g.OwnerID)
	txt := fmt.Sprintf("Serveur quitté : %s (Owner : %s (ID: %s)) avec %d membres.", g.Name, owner.Username+"#"+owner.Discriminator, g.OwnerID, len(g.Members))
	Log("Système | Avertissement", txt)
	LogFile("Système | Avertissement", "", txt)
}

func registerCommands(c *CommandHandler) {
	c.AddCommand("ping", "Système", nil, "Réponds par pong si le bot est en ligne", PingCommand, false, false)
	c.AddCommand("help", "Informations", nil, "Affiche la liste des commandes", HelpCommand, false, false)
	c.AddCommand("create", "RolePlay", nil, "Crée le joueur dans la BDD", PlayerCreate, false, false)
	c.AddCommand("display", "RolePlay", nil, "Affiche les infos sur l'id/la mention donnée", DisplayPlayer, false, false)
	c.AddCommand("code", "RolePlay", nil, "Moyen de gagner des bits", CodeCommand, false, false)
	c.AddCommand("exec-sql", "Système", []string{"sql-exec"}, "Exécute le code SQL donné", ExecSQLCommand, false, true)
	c.AddCommand("shutdown", "Système", []string{"close", "stop", "kill"}, "Eteint le bot proprement", ShutdownCommand, false, true)
	c.AddCommand("buy", "RolePlay", []string{"shop"}, "Vous ajoute le skill demandé / affiche le shop", BuyCommand, false, false)
	fmt.Println()
}

func readConfig() {
	jsonF, err := os.Open("data/config.json")

	if err != nil {
		log.Fatal("Erreur en lisant [config.json]")
	}
	defer jsonF.Close()

	byteVal, _ := ioutil.ReadAll(jsonF)
	json.Unmarshal(byteVal, &Config)

	Config.Version = "v0.0.1"
	if runtime.GOOS == "windows" {
		Config.Prefix = "dv;"
	}

	registerCommands(&Config.CommandHandler)
}

func isInCommandsCooldown(id string) bool {
	for _, i := range commandsCooldown {
		if i == id {
			return true
		}
	}
	return false
}
