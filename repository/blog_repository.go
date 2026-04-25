/*
这个文件实现博客数据的数据访问逻辑
*/
package repository

import (
	"blog/model"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// BlogRepository 负责读取和写入博客数据
type BlogRepository struct {
	db *sql.DB
}

// NewBlogRepository 创建博客仓储实例
func NewBlogRepository(db *sql.DB) *BlogRepository {
	return &BlogRepository{db: db}
}

// List 查询博客列表 支持分页和关键字搜索
func (r *BlogRepository) List(page, pageSize int, keyword string) (*model.BlogListResult, error) {
	keyword = strings.TrimSpace(keyword)
	offset := (page - 1) * pageSize
	likeKeyword := "%" + keyword + "%"

	countQuery := `
		SELECT COUNT(*)
		FROM posts p
		LEFT JOIN users u ON u.id = p.author_id
	`
	listQuery := `
		SELECT p.id, p.author_id, p.title, COALESCE(p.slug, ''), COALESCE(p.summary, ''), COALESCE(pc.content_markdown, ''), COALESCE(u.username, ''), p.created_at
		FROM posts p
		LEFT JOIN users u ON u.id = p.author_id
		LEFT JOIN post_contents pc ON pc.post_id = p.id
	`

	var countArgs []any
	var listArgs []any
	baseFilter := `
		WHERE p.deleted_at IS NULL AND p.status = 'published'
	`
	countQuery += baseFilter
	listQuery += baseFilter
	if keyword != "" {
		filter := `
			AND (p.title LIKE ? OR COALESCE(p.summary, '') LIKE ? OR COALESCE(pc.content_text, pc.content_markdown, '') LIKE ? OR COALESCE(u.username, '') LIKE ?)
		`
		countQuery += filter
		listQuery += filter
		countArgs = append(countArgs, likeKeyword, likeKeyword, likeKeyword, likeKeyword)
		listArgs = append(listArgs, likeKeyword, likeKeyword, likeKeyword, likeKeyword)
	}

	var total int
	if err := r.db.QueryRow(countQuery, countArgs...).Scan(&total); err != nil {
		return nil, err
	}

	listQuery += `
		ORDER BY p.published_at DESC, p.id DESC
		LIMIT ? OFFSET ?
	`
	listArgs = append(listArgs, pageSize, offset)

	rows, err := r.db.Query(listQuery, listArgs...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var blogs []model.Blog
	for rows.Next() {
		var blog model.Blog

		// 列表接口把博客标识 作者和创建时间一起带给前端
		// 这样详情跳转 权限判断和展示都不需要再猜测来源
		if err := rows.Scan(&blog.ID, &blog.AuthorID, &blog.Title, &blog.Slug, &blog.Summary, &blog.Content, &blog.AuthorUsername, &blog.CreatedAt); err != nil {
			return nil, err
		}
		blogs = append(blogs, blog)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &model.BlogListResult{
		Items:    blogs,
		Page:     page,
		PageSize: pageSize,
		Total:    total,
		Keyword:  keyword,
	}, nil
}

// Create 创建新博客
func (r *BlogRepository) Create(blog *model.Blog) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var authorID int64
	if err := tx.QueryRow("SELECT id FROM users WHERE username=? AND deleted_at IS NULL", blog.AuthorUsername).Scan(&authorID); err != nil {
		return err
	}

	slug := blog.Slug
	if slug == "" {
		slug = fmt.Sprintf("post-%d", time.Now().UnixNano())
	}

	result, err := tx.Exec(`
		INSERT INTO posts (author_id, title, slug, summary, status, visibility, published_at)
		VALUES (?, ?, ?, ?, 'published', 'public', NOW())
	`, authorID, blog.Title, slug, blog.Summary)
	if err != nil {
		return err
	}

	postID, err := result.LastInsertId()
	if err != nil {
		return err
	}

	if _, err := tx.Exec(`
		INSERT INTO post_contents (post_id, content_markdown, content_text, word_count)
		VALUES (?, ?, ?, ?)
	`, postID, blog.Content, blog.Content, len([]rune(blog.Content))); err != nil {
		return err
	}

	if _, err := tx.Exec("INSERT INTO post_stats (post_id) VALUES (?)", postID); err != nil {
		return err
	}

	return tx.Commit()
}

// GetByID 查询指定博客的完整信息
func (r *BlogRepository) GetByID(blogID int64) (*model.Blog, error) {
	var blog model.Blog

	err := r.db.QueryRow(`
		SELECT p.id, p.author_id, p.title, COALESCE(p.slug, ''), COALESCE(p.summary, ''), COALESCE(pc.content_markdown, ''), COALESCE(u.username, ''), p.created_at
		FROM posts p
		LEFT JOIN users u ON u.id = p.author_id
		LEFT JOIN post_contents pc ON pc.post_id = p.id
		WHERE p.id=? AND p.deleted_at IS NULL
	`, blogID).Scan(&blog.ID, &blog.AuthorID, &blog.Title, &blog.Slug, &blog.Summary, &blog.Content, &blog.AuthorUsername, &blog.CreatedAt)
	if err != nil {
		return nil, err
	}

	return &blog, nil
}

// GetAuthorByID 查询指定博客的作者用户名
func (r *BlogRepository) GetAuthorByID(blogID int64) (string, error) {
	var authorUsername string
	err := r.db.QueryRow(`
		SELECT COALESCE(u.username, '')
		FROM posts p
		LEFT JOIN users u ON u.id = p.author_id
		WHERE p.id=? AND p.deleted_at IS NULL
	`, blogID).Scan(&authorUsername)
	if err != nil {
		return "", err
	}
	return authorUsername, nil
}

// Update 更新指定博客的标题和内容
func (r *BlogRepository) Update(blog *model.Blog) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`
		UPDATE posts
		SET title=?, summary=?, updated_at=NOW()
		WHERE id=? AND deleted_at IS NULL
	`, blog.Title, blog.Summary, blog.ID); err != nil {
		return err
	}

	if _, err := tx.Exec(`
		UPDATE post_contents
		SET content_markdown=?, content_text=?, word_count=?, updated_at=NOW()
		WHERE post_id=?
	`, blog.Content, blog.Content, len([]rune(blog.Content)), blog.ID); err != nil {
		return err
	}

	return tx.Commit()
}

// Delete 删除指定博客
func (r *BlogRepository) Delete(blogID int64) error {
	_, err := r.db.Exec("DELETE FROM posts WHERE id=?", blogID)
	return err
}
