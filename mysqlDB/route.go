package mysqlDB

import (
	"blog/redisDB"
	"database/sql"
	"html/template"
	"log"

	"github.com/gin-gonic/gin"
)

func RegisterRouters(r *gin.Engine, db *sql.DB) {
	//主页
	//r.GET("/", func(c *gin.Context) {
	//	blogs, err := GetBlogs(db)
	//	if err != nil {
	//		log.Fatal("Error fetching blogs:", err)
	//	}
	//	c.HTML(200, "index.html", gin.H{
	//		"title": "Xuan",
	//		"blogs": blogs,
	//	})
	//
	//})
	r.GET("/", IndexHandler(db))

	//博客提交模块
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

	//注册模块
	r.GET("/register", func(c *gin.Context) {
		c.HTML(200, "register.html", gin.H{
			"title": "Register",
		})
	})
	r.POST("/register", func(c *gin.Context) {
		username := c.PostForm("username")
		password := c.PostForm("password")

		success := registerNewUser(db, username, password)

		if success == "Register success" {
			c.JSON(200, gin.H{
				"message": "Register success",
			})
		} else {
			c.JSON(400, gin.H{
				"message": success,
			})
		}
	})

	//登录模块
	r.GET("/login", func(c *gin.Context) {
		c.HTML(200, "login.html", gin.H{
			"title": "Login",
		})
	})
	r.POST("/login", func(c *gin.Context) {
		username := c.PostForm("username")
		password := c.PostForm("password")

		success := LoginUser(db, username, password)

		if success == "Success Login" {
			sessionID, _ := redisDB.CreateSession(username)
			c.SetCookie("sessionID", sessionID, 3600, "/", "", false, true)
			c.JSON(200, gin.H{
				"message": "Login success",
			})
		} else {
			c.JSON(400, gin.H{
				"message": success,
			})
		}
	})

	if err := r.Run(":5345"); err != nil {
		log.Fatal(err)
	}
}

// 登录
func RegisterAuthRouters(r *gin.RouterGroup, db *sql.DB) {

	r.GET("/user", func(c *gin.Context) {

		userID := c.GetString("userID")

		c.JSON(200, gin.H{
			"message": "welcome",
			"user":    userID,
		})
	})
}

func LoadTemplates() *template.Template {
	funcMap := template.FuncMap{
		"blog_title":   func(b Blog) string { return b.Title },
		"blog_content": func(b Blog) string { return b.Content },
	}
	return template.Must(template.New("templates").Funcs(funcMap).ParseFiles(
		"templates/index.html",
		"templates/register.html",
		"templates/login.html",
	))
}
