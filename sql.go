package main

import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
)

type Blog struct {
	Title   string
	Content string
}

func InitDB() (*sql.DB, error) {
	dsn := "blog:123456@tcp(127.0.0.1:3306)/blog?charset=utf8"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func getBlogs(db *sql.DB) ([]Blog, error) {
	rows, err := db.Query("SELECT blog_title, blog_content FROM blog")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var blogs []Blog
	for rows.Next() {
		var blog Blog
		if err := rows.Scan(&blog.Title, &blog.Content); err != nil {
			return nil, err
		}
		blogs = append(blogs, blog)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return blogs, nil
}
