package main

import (
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
