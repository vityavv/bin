package main

import (
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
	"database/sql"
	"errors"
)

var db *sql.DB

func DBinit() error {
	var err error
	db, err = sql.Open("sqlite3", "file:./database.sqlite?cache=shared&mode=rwc")
	if err != nil {
		return err
	}
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS users (id INTEGER PRIMARY KEY AUTOINCREMENT, username TEXT UNIQUE, password TEXT)")
	return err
}
func DBclose() {db.Close()}

// User methods: {{{
func DBcreateUser(username, password string) error {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	_, err = db.Exec("INSERT INTO users (username, password) VALUES (?, ?)", username, string(hashed))
	if err != nil {
		return err
	}
	return nil
}
func DBlogIn(username, password string) (bool, error) {
	rows, err := db.Query("SELECT password FROM users WHERE username = ?", username)
	if err != nil {
		return false, err
	}
	defer rows.Close()
	if rows.Next() {
		var hashed string
		err = rows.Scan(&hashed)
		if err != nil {
			return false, err
		}
		err = bcrypt.CompareHashAndPassword([]byte(hashed), []byte(password))
		if err != nil {
			return false, nil
		}
		return true, nil
	}
	return false, errors.New("User not found")
}
// }}}
