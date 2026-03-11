package mysqlDB

import (
	"database/sql"
	"log"

	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
)

func registerNewUser(db *sql.DB, username string, password string) string {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Println("Error hashing password:", err)
		return "Register failed"
	}

	storedUsernameQuery := "SELECT username FROM user WHERE username=?"
	var storedUsername string
	err = db.QueryRow(storedUsernameQuery, username).Scan(&storedUsername)
	if err != nil {
		log.Println(err)
	}
	if storedUsername == username {
		log.Println("User already exists")
		return "User already exists"
	}

	registerQuery := "INSERT INTO user (username, password) VALUES (?, ?)"
	_, err = db.Exec(registerQuery, username, hashedPassword)
	if err != nil {
		log.Println("Error inserting new user", err)
		return "Register failed"
	}

	return "Register success"
}
