package main

import (
	"database/sql"
	_ "database/sql"
	"html/template"
	"log"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	r := gin.Default()

	dsn := "blog:123456@tcp(127.0.0.1:3306)/blog?charset=utf8"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("Error connecting to database:", err)
	}
	defer db.Close()

	//定义模板函数
	funcMap := template.FuncMap{
		"blog_title": func() string {
			var title string
			err := db.QueryRow("select blog_title from blog where blog_id = 1").Scan(&title)
			if err != nil {
				log.Fatal(err)
			}
			return title
		},
		"blog_content": func() string {
			var content string
			err := db.QueryRow("select blog_content from blog where blog_id = 1").Scan(&content)
			if err != nil {
				log.Fatal(err)
			}
			return content
		},
	}

	//加载模板并传递函数映射
	r.SetHTMLTemplate(template.Must(template.New("templates").Funcs(funcMap).ParseFiles("templates/index.html")))

	//r.LoadHTMLGlob("templates/*.html") //比较老的没那么好用的加载方式
	r.Static("/css", "./templates/css")
	r.Static("/js", "./templates/js")
	r.Static("/img", "./templates/img")

	r.GET("/", func(c *gin.Context) {
		c.HTML(200, "index.html", gin.H{
			"title": "Xuan",
		})
	})

	r.GET("/submit", func(c *gin.Context) {
		c.JSON(400, gin.H{
			"400": "Access method denied",
		})
	})

	r.POST("/submit", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "POST test",
		})
	})

	if err := r.Run(":5345"); err != nil {
		log.Fatal(err)
	}
}

//func MiddleWare() gin.HandlerFunc {
//	r := gin.New()
//
//	// 使用日志中间件
//	logger := zap.New()
//	r.Use(ginzap.Ginzap(logger, time.RFC3339, true))
//
//	// ... 其他路由和处理函数
//
//	r.Run(":8080")
//}
