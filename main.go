package main

import (
	"blog/mysqlDB"
	"blog/redisDB"
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

	//连接mysql
	db, err := mysqlDB.InitDB()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	//连接Redis
	err = redisDB.InitRedis()
	if err != nil {
		log.Fatal(err)
	}

	//加载模板
	r.SetHTMLTemplate(mysqlDB.LoadTemplates())

	//公共路由
	mysqlDB.RegisterRouters(r, db)

	//登录验证
	auth := r.Group("/")
	auth.Use(redisDB.AuthMiddleware())
	mysqlDB.RegisterAuthRouters(auth, db)

	if err := r.Run(":5345"); err != nil {
		log.Fatal(err)
	}
}
