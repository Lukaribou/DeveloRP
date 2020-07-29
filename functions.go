package main

import (
	"fmt"
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
	fmt.Printf("[%s] | [%s] %s", TimeFormatFr(time.Now()), tag, fmt.Sprintf(msg, a...))
}
