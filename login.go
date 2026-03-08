package main

import (
	"database/sql"
	"log"

	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
)

// 未加密密码,暂时以明文存储
func LoginUser(db *sql.DB, username string, password string) string {

	loginQuery := "select password from user where username=? "

	var storedPassword string
	err := db.QueryRow(loginQuery, username).Scan(&storedPassword)
	if err != nil { //用户不存在或查询出错
		log.Println("Error querying user:", err)
		return "Username or Password is Invalid"
	}

	err = bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(password))
	if err != nil {
		//密码不匹配
		log.Println("password is wrong", err)
		return "Username or Password is Invalid"
	}
	return "Success Login"
}
