package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

var (
	config Config
)

func main() {
	config := readConfig()
	dg, err := discordgo.New("Bot " + config.Token)
	if err != nil {
		log.Fatal("Erreur lors de la création du client")
	}

	dg.AddHandler(messageCreate)

	err = dg.Open()
	if err != nil {
		log.Fatal("Erreur lors de l'ouverture de la connection")
	}
	defer dg.Close()

	fmt.Println("Bot connecté !")

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
	} else if !strings.HasPrefix(m.Content, config.Prefix) {
		return
	}
	m.Content = m.Content[3:]

	args := strings.Split(strings.ToLower(m.Content), " ")
	cmd, err := config.CommandHandler.Get(args[0])
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "**La commande `"+args[0]+"` m'est inconnue.**")
		return
	}

	if cmd.OwnerOnly && config.OwnerID != m.Author.ID {
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
	c, _ := strconv.Atoi(m.ChannelID)

	cmd.Execute(Context{g, g.Channels[c], m.Member, m, s, args})
}

func registerCommands() []Command {
	var cmds []Command
	cmds = append(cmds, Command{"ping", []string{}, "Réponds par pong si le bot est en ligne", false, false, pingCommand})

	return cmds
}

func readConfig() Config {
	jsonF, err := os.Open("data/config.json")

	if err != nil {
		log.Fatal("Erreur en lisant [config.json]")
	}
	defer jsonF.Close()

	byteVal, _ := ioutil.ReadAll(jsonF)

	var cfg Config
	json.Unmarshal(byteVal, &cfg)
	cfg.CommandHandler = CommandHandler{registerCommands()}

	return cfg
}
