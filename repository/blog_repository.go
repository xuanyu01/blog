/*
该文件实现博客数据的数据库访问逻辑
*/
package repository

import (
	"blog/model"
	"database/sql"
)

// BlogRepository 负责读取博客数据
// 它封装与博客列表查询相关的 SQL 访问细节
type BlogRepository struct {
	db *sql.DB
}

// NewBlogRepository 创建博客仓储实例
func NewBlogRepository(db *sql.DB) *BlogRepository {
	return &BlogRepository{db: db}
}

// List 查询博客列表
func (r *BlogRepository) List() ([]model.Blog, error) {
	rows, err := r.db.Query("SELECT blog_title, blog_content FROM blog")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var blogs []model.Blog
	for rows.Next() {
		var blog model.Blog

		// 逐行扫描结果，保持数据库字段到领域模型的映射集中在仓储层
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
