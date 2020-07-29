package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
)

var (
	// Config : La config du bot
	Config ConfigStruct
)

func main() {
	readConfig()
	dg, err := discordgo.New("Bot " + Config.Token)
	if err != nil {
		log.Fatal("Erreur lors de la création du client")
	}

	dg.AddHandler(messageCreate)

	err = dg.Open()
	if err != nil {
		log.Fatal("Erreur lors de l'ouverture de la connection")
	}
	defer dg.Close()

	rand.Seed(time.Now().UnixNano()) // Initialiser le rand

	Log("Système", "Bot en ligne sur %d serveurs sous le nom de %s.\n", len(dg.State.Guilds), dg.State.User.Username)

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.Bot {
		return
	} else if c, _ := s.Channel(m.ChannelID); c.Type == discordgo.ChannelTypeDM || c.Type == discordgo.ChannelTypeGroupDM {
		s.ChannelMessageSend(c.ID, "**Je ne prends pas en compte les commandes effectuées en MP.**")
		return
	} else if !strings.HasPrefix(m.Content, Config.Prefix) {
		return
	}
	m.Content = m.Content[len(Config.Prefix):]

	args := strings.Split(strings.ToLower(m.Content), " ")
	cmd, err := Config.CommandHandler.Get(args[0])
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "**La commande `"+args[0]+"` m'est inconnue.**")
		return
	}

	if cmd.OwnerOnly && Config.OwnerID != m.Author.ID {
		s.ChannelMessageSend(m.ChannelID, "**Cette commande est limitée au propriétaire du bot.**")
		return
	} else if p, _ := MemberHasPermission(s, m.GuildID, m.Author.ID, discordgo.PermissionAdministrator); cmd.GuildAdminsOnly && !p {
		s.ChannelMessageSend(m.ChannelID, "**Cette commande est réservée aux administrateurs du serveur.**")
		return
	}
	g, err := s.Guild(m.GuildID)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "**Une erreur est survenue.**")
		return
	}

	if c, err := s.Channel(m.ChannelID); err == nil {
		go cmd.Execute(Context{g, c, m.Member, m, s, args})
	}
}

func registerCommands(c *CommandHandler) {
	c.AddCommand("ping", nil, "Réponds par pong si le bot est en ligne", pingCommand, false, false)
	c.AddCommand("help", nil, "Affiche la liste des commandes", helpCommand, false, false)
}

func readConfig() {
	jsonF, err := os.Open("data/config.json")

	if err != nil {
		log.Fatal("Erreur en lisant [config.json]")
	}
	defer jsonF.Close()

	byteVal, _ := ioutil.ReadAll(jsonF)
	json.Unmarshal(byteVal, &Config)

	registerCommands(&Config.CommandHandler)
}
