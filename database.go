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
	Log("BDD", "Connexion réussie / Version: %s", v)
	return d
}

// PlayerExist : Renvoie true si l'id se trouve dans la bdd
func (db *DB) PlayerExist(userID string) bool {
	var u string
	e := db.sql.QueryRow("SELECT ID FROM users WHERE userID = ?", userID).Scan(&u)
	return e != sql.ErrNoRows && e == nil
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
		&pl.lastCode)

	return &pl, nil
}
