package main

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/fatih/color"
)

// ArrIncludes : Indique si un tableau contient une certaine chaine
func ArrIncludes(arr []string, str string) bool {
	for _, s := range arr {
		if s == str {
			return true
		}
	}
	return false
}

// MemberHasPermission : Vérifie que le membre a bien les permissions nécessaires
// https://github.com/bwmarrin/discordgo/wiki/FAQ#permissions-and-roles
func MemberHasPermission(s *discordgo.Session, guildID string, userID string, permission int) (bool, error) {
	member, err := s.State.Member(guildID, userID)
	if err != nil {
		if member, err = s.GuildMember(guildID, userID); err != nil {
			return false, err
		}
	}

	// Iterate through the role IDs stored in member.Roles
	// to check permissions
	for _, roleID := range member.Roles {
		role, err := s.State.Role(guildID, roleID)
		if err != nil {
			return false, err
		}
		if role.Permissions&permission != 0 {
			return true, nil
		}
	}

	return false, nil
}

// TimeFormatFr : Transforme un time.Time en date formatée pour le FR
func TimeFormatFr(time time.Time) string {
	return time.Format("02/01/2006 15h04m05")
}

// TimestampSecToDate : Renvoie la date correspondant au timestamp secondes
func TimestampSecToDate(sec int64) string {
	return TimeFormatFr(time.Unix(sec, 0))
}

// TimestampNanoToDate : Renvoie la date correspondant au timestamp nanos
func TimestampNanoToDate(nano int64) string {
	return TimeFormatFr(time.Unix(0, nano))
}

func formatLogTag(tag string) string {
	switch tag {
	case "Err":
		tag = "Erreur"
	case "Sys":
		tag = "Système"
	case "Sys S":
		tag = "Système | Succès"
	case "Sys Err":
		tag = "Système | Erreur"
	case "BDD":
		tag = "Base de données"
	case "BDD S":
		tag = "Base de données | Succès"
	case "BDD Err":
		tag = "Base de données | Erreur"
	case "Warn":
		tag = "Avertissement"
	case "S":
		tag = "Succès"
	default:
		//
	}
	return tag
}

// Log : Printf mais formaté pour la console
func Log(tag, msg string, a ...interface{}) {
	tag = formatLogTag(tag)
	text := fmt.Sprintf("[%s] | [%s] %s\n", TimeFormatFr(time.Now()), tag, fmt.Sprintf(msg, a...))
	if strings.Contains(tag, "Erreur") {
		color.HiRed(text)
	} else if strings.Contains(tag, "Avertissement") {
		color.HiYellow(text)
	} else if strings.Contains(tag, "Succès") {
		color.HiGreen(text)
	} else {
		fmt.Print(text)
	}
}

// LogFile : Log() mais vers un fichier
func LogFile(tag, fileName, msg string, a ...interface{}) {
	tag = formatLogTag(tag)
	if fileName == "" {
		fileName = "data/log.txt"
	}

	f, err := os.OpenFile(fileName, os.O_APPEND, 0666)
	if err != nil {
		if os.IsNotExist(err) {
			cf, e := os.Create(fileName)
			if e != nil {
				Log("Sys Err", "Erreur LogFile : %s", err.Error())
				return
			}
			f = cf
			if _, er := f.WriteString("[ Logs DeveloRP " + Config.Version + " ]\n\n"); er != nil {
				Log("Sys Err", "Erreur LogFile : %s", err.Error())
				return
			}
		} else {
			Log("Sys Err", "Erreur LogFile : %s", err.Error())
			return
		}
	}

	defer f.Close()
	if _, e := f.WriteString(fmt.Sprintf("[%s] | [%s] %s\n", TimeFormatFr(time.Now()), tag, fmt.Sprintf(msg, a...))); e != nil {
		Log("Sys Err", "Erreur LogFile : %s", err.Error())
	}
}

// InPercentLuck : ...
func InPercentLuck(i int) bool {
	return i < rand.Intn(101) // (rand.Intn(max - min + 1) + min)
}

// RandomInt : Génère un nombre aléatoire en min et max
func RandomInt(min, max int) int {
	return rand.Intn(max-min+1) + min
}

// GetEmojiOkOrX : Renvoie l'émoji Check si la condition == true, sinon X
func GetEmojiOkOrX(cond bool) string {
	if cond {
		return OKEMOJI
	}
	return XEMOJI
}

// RandomColor : Renvoie une couleur aléatoire en hexa
func RandomColor() int {
	l := []string{"A", "B", "C", "D", "E", "F", "0", "1", "2", "3", "4", "5", "6", "7", "8", "9"}
	var s string
	for i := 0; i < 6; i++ {
		s = s + l[rand.Intn(16)]
	}
	r, _ := strconv.ParseInt(s, 16, 0)
	return int(r)
}

// ArrayFind : Retourne la position d'un élément dans le tableau
// Retourne -1 si l'élément n'est pas dans le tableau
func ArrayFind(a []string, x string) int {
	for i, n := range a {
		if x == n {
			return i
		}
	}
	return -1
}

// ArrayRemove : Retire l'élément à l'index i
func ArrayRemove(a []string, i int) []string {
	a[i] = a[len(a)-1]
	// We do not need to put s[i] at the end, as it will be discarded anyway
	return a[:len(a)-1]
}
