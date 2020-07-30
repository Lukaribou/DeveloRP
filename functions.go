package main

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/bwmarrin/discordgo"
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

// TimestampToDate : Renvoie la date correspondant au timestamp
func TimestampToDate(nano int64) string {
	return TimeFormatFr(time.Unix(0, nano))
}

// Log : Printf mais formaté pour la console
func Log(tag string, msg string, a ...interface{}) {
	fmt.Printf("[%s] | [%s] %s\n", TimeFormatFr(time.Now()), tag, fmt.Sprintf(msg, a...))
}

// InPercentLuck : ...
func InPercentLuck(i int) bool {
	return i < rand.Intn(101) // (rand.Intn(max - min + 1) + min)
}

// RandomInt : Génère un nombre aléatoire en min et max
func RandomInt(min, max int) int {
	return rand.Intn(max-min+1) + min
}

// Constantes des émojis
const (
	OKEMOJI        = "✅"
	XEMOJI         = "❌"
	WARNINGEMOJI   = "⚠"
	RIGHTARROW     = "➡"
	TADAEMOJI      = "🎉"
	ADMINSEMOJI    = "🚔"
	OWNERONLYEMOJI = "🔐"
)

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
