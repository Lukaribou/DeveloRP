package main

import (
	"database/sql"
	"errors"
	"runtime"

	_ "github.com/go-sql-driver/mysql"
)

// DB : ...
type DB struct {
	sql *sql.DB
}

// NewDB : Initialise la BDD
func NewDB() *DB {
	return &DB{
		sql: DbConnect(),
	}
}

// DbConnect : Se connecte à la base de données
func DbConnect() *sql.DB {
	var sqlConn string
	if runtime.GOOS == "windows" {
		sqlConn = "root:" + Config.DbPassword + "@tcp(:" + Config.SQLPort[0] + ")/develorp"
	} else {
		sqlConn = "pi:" + Config.DbPassword + "@tcp(:" + Config.SQLPort[1] + ")/develorp"
	}

	d, err := sql.Open("mysql", sqlConn)
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

	pl.db = db
	return &pl, nil
}

// ExistLanguage : Retourne true si le language est dans la BDD
func (db *DB) ExistLanguage(name string) bool {
	return db.ExecExistWithQuery("SELECT ID FROM langs WHERE name = ?", name)
}

// GetLanguage : Prends le langage correspondant à l'ID donné
func (db *DB) GetLanguage(langName string) (*Language, error) {
	if !db.ExecExistWithQuery("SELECT ID FROM langs WHERE name = ?", langName) {
		return nil, errors.New("Le '" + langName + "' ne correspond à aucun langage.")
	}
	var l Language
	db.sql.QueryRow("SELECT * FROM langs WHERE name = ?", langName).Scan(
		&l.ID,
		&l.name,
		&l.level,
		&l.skills,
		&l.cost)

	l.db = db

	return &l, nil
}

// GetSkills : Renvoie la liste de tous les skills de la BDD
func (db *DB) GetSkills() []*Skill {
	var skills []*Skill
	rows, err := db.sql.Query("SELECT * FROM skills")
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	for rows.Next() {
		var s Skill
		err := rows.Scan(&s.ID, &s.cost, &s.gain, &s.name)
		if err != nil {
			panic(err)
		}
		s.db = db
		skills = append(skills, &s)
	}
	return skills
}
