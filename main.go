package main

import (
	_ "database/sql"
	"log"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	r := gin.Default()

	//加载模板并传递函数映射

	r.Static("/css", "./templates/css")
	r.Static("/js", "./templates/js")
	r.Static("/img", "./templates/img")

	db, err := InitDB()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	r.SetHTMLTemplate(LoadTemplates())

	RegisterRouters(r, db)

	if err := r.Run(":5345"); err != nil {
		log.Fatal(err)
	}
}
