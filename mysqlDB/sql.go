package mysqlDB

import (
	"database/sql"
	"fmt"
	"time"

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

	//检查数据库连接
	if err := db.Ping(); err != nil {
		return nil, err
	}

	//连接池
	db.SetConnMaxLifetime(time.Hour)
	db.SetMaxIdleConns(10)
	db.SetMaxOpenConns(20)

	fmt.Println("mysql connect success")

	return db, nil
}

func GetBlogs(db *sql.DB) ([]Blog, error) {
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

func GetUserImage(db *sql.DB, username string) (string, error) {

	var image string

	err := db.QueryRow(
		"SELECT image FROM user WHERE username=?",
		username,
	).Scan(&image)

	if err != nil {
		return "", err
	}

	return image, nil
}
