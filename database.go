package main

import (
	"database/sql"
	"errors"

	_ "github.com/go-sql-driver/mysql"
)

// DB : ...
type DB struct {
	sql *sql.DB
}

// NewDB : Initialise la BDD
func NewDB() *DB {
	d := &DB{}
	return &DB{sql: d.DbConnect()}
}

// DbConnect : Se connecte à la base de données
func (db *DB) DbConnect() *sql.DB {
	d, err := sql.Open("mysql", "root:"+Config.DbPassword+"@/develorp")
	if err != nil {
		panic(err)
	}
	var v string
	e := d.QueryRow("SELECT VERSION()").Scan(&v)
	if e != nil {
		panic(e)
	}
	Log("BDD", "Connexion réussie / Version: %s.", v)
	return d
}

// ExecExistWithQuery : ...
func (db *DB) ExecExistWithQuery(tmpl string, el ...interface{}) bool {
	var u string
	e := db.sql.QueryRow(tmpl, el...).Scan(&u)
	return e == nil
}

// PlayerExist : Renvoie true si l'id se trouve dans la bdd
func (db *DB) PlayerExist(userID string) bool {
	return db.ExecExistWithQuery("SELECT ID FROM users WHERE userID = ?", userID)
}

// GetPlayer : Renvoie l'utilisateur si il existe, une erreur sinon
func (db *DB) GetPlayer(userID string) (*Player, error) {
	if !db.PlayerExist(userID) {
		return nil, errors.New("L'utilisateur " + userID + " n'existe pas dans la base de données.")
	}
	var pl Player
	db.sql.QueryRow("SELECT * FROM users WHERE userID = ?", userID).Scan(
		&pl.ID,
		&pl.userID,
		&pl.money,
		&pl.level,
		&pl.createDate,
		&pl.lastCode,
		&pl.skills)

	return &pl, nil
}

// ExistLanguage : Retourne true si le language est dans la BDD
func (db *DB) ExistLanguage(name string) bool {
	return db.ExecExistWithQuery("SELECT ID FROM langs WHERE name = ?", name)
}

// GetCurrentLanguage : Prends le langage correspondant à l'ID donné
func (db *DB) GetCurrentLanguage(curLangName string) (*Language, error) {
	if !db.ExecExistWithQuery("SELECT ID FROM langs WHERE name = ?", curLangName) {
		return nil, errors.New("Le '" + curLangName + "' ne correspond à aucun langage.")
	}
	var l Language
	db.sql.QueryRow("SELECT * FROM langs WHERE name = ?", curLangName).Scan(
		&l.ID,
		&l.name,
		&l.level,
		&l.skills)

	return &l, nil
}
