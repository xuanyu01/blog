package main

import (
	"database/sql"
	"html/template"
	"log"

	"github.com/gin-gonic/gin"
)

func RegisterRouters(r *gin.Engine, db *sql.DB) {
	r.GET("/", func(c *gin.Context) {
		blogs, err := getBlogs(db)
		if err != nil {
			log.Fatal("Error fetching blogs:", err)
		}
		c.HTML(200, "index.html", gin.H{
			"title": "Xuan",
			"blogs": blogs,
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

func LoadTemplates() *template.Template {
	funcMap := template.FuncMap{
		"blog_title":   func(b Blog) string { return b.Title },
		"blog_content": func(b Blog) string { return b.Content },
	}
	return template.Must(template.New("templates").Funcs(funcMap).ParseFiles("templates/index.html"))
}
