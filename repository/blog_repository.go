/*
这个文件实现博客数据的数据访问逻辑
*/
package repository

import (
	"blog/model"
	"database/sql"
)

// BlogRepository 负责读取和写入博客数据
type BlogRepository struct {
	db *sql.DB
}

// NewBlogRepository 创建博客仓储实例
func NewBlogRepository(db *sql.DB) *BlogRepository {
	return &BlogRepository{db: db}
}

// List 查询博客列表
func (r *BlogRepository) List() ([]model.Blog, error) {
	rows, err := r.db.Query(`
		SELECT blog_id, blog_title, blog_content, COALESCE(author_username, ''), created_at
		FROM blog
		ORDER BY created_at DESC, blog_id DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var blogs []model.Blog
	for rows.Next() {
		var blog model.Blog

		// 列表接口把博客标识 作者和创建时间一起带给前端
		// 这样详情跳转 权限判断和展示都不需要再猜测来源
		if err := rows.Scan(&blog.ID, &blog.Title, &blog.Content, &blog.AuthorUsername, &blog.CreatedAt); err != nil {
			return nil, err
		}
		blogs = append(blogs, blog)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return blogs, nil
}

// Create 创建新博客
func (r *BlogRepository) Create(blog *model.Blog) error {
	_, err := r.db.Exec(
		"INSERT INTO blog (blog_title, blog_content, author_username) VALUES (?, ?, ?)",
		blog.Title,
		blog.Content,
		blog.AuthorUsername,
	)
	return err
}

// GetAuthorByID 查询指定博客的作者用户名
func (r *BlogRepository) GetAuthorByID(blogID int64) (string, error) {
	var authorUsername string
	err := r.db.QueryRow("SELECT COALESCE(author_username, '') FROM blog WHERE blog_id=?", blogID).Scan(&authorUsername)
	if err != nil {
		return "", err
	}
	return authorUsername, nil
}

// Delete 删除指定博客
func (r *BlogRepository) Delete(blogID int64) error {
	_, err := r.db.Exec("DELETE FROM blog WHERE blog_id=?", blogID)
	return err
}
