/*
实现博客数据访问逻辑。
*/
package repository

import (
	"blog/model"
	"database/sql"
	"fmt"
	"gorm.io/gorm"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// BlogRepository 负责博客数据读写。
type BlogRepository struct {
	db *gorm.DB
}

// NewBlogRepository 创建博客仓储。
func NewBlogRepository(db *gorm.DB) *BlogRepository {
	return &BlogRepository{db: db}
}

// List 返回前台博客列表。
func (r *BlogRepository) List(page, pageSize int, query model.BlogListQuery) (*model.BlogListResult, error) {
	offset := (page - 1) * pageSize
	keyword := strings.TrimSpace(query.Keyword)
	archive := strings.TrimSpace(query.Archive)

	countQuery := `
		SELECT COUNT(*)
		FROM posts p
		LEFT JOIN users u ON u.id = p.author_id
		LEFT JOIN post_contents pc ON pc.post_id = p.id
		WHERE p.deleted_at IS NULL AND p.status = 'published'
	`
	listQuery := `
		SELECT
			p.id,
			p.author_id,
			COALESCE(p.category_id, 0),
			COALESCE(c.name, ''),
			COALESCE(c.slug, ''),
			p.title,
			COALESCE(p.slug, ''),
			COALESCE(p.summary, ''),
			COALESCE(pc.content_markdown, ''),
			COALESCE(u.username, ''),
			p.status,
			p.is_top,
			p.created_at,
			p.updated_at,
			p.published_at,
			COALESCE(ps.view_count, 0),
			COALESCE(ps.like_count, 0),
			COALESCE(ps.favorite_count, 0),
			COALESCE(ps.comment_count, 0),
			COALESCE(GROUP_CONCAT(DISTINCT CONCAT(t.id, '::', t.name, '::', t.slug) ORDER BY t.name SEPARATOR '||'), '')
	`

	countArgs, listArgs := buildBlogListFilters(keyword, query.CategoryID, archive)
	countQuery += countArgs.query
	if keyword != "" {
		listQuery += `,
			MAX(CASE
				WHEN COALESCE(t.name, '') LIKE ? OR COALESCE(t.slug, '') LIKE ? THEN 1
				ELSE 0
			END) AS tag_match_score,
			MAX(CASE
				WHEN p.title LIKE ? THEN 2
				WHEN COALESCE(p.summary, '') LIKE ? THEN 1
				WHEN COALESCE(pc.content_text, pc.content_markdown, '') LIKE ? THEN 0
				ELSE -1
			END) AS text_match_score
		`
		listArgs.args = append(listArgs.args,
			"%"+keyword+"%", "%"+keyword+"%",
			"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%",
		)
	} else {
		listQuery += `,
			0 AS tag_match_score,
			0 AS text_match_score
		`
	}
	listQuery += `
		FROM posts p
		LEFT JOIN users u ON u.id = p.author_id
		LEFT JOIN categories c ON c.id = p.category_id
		LEFT JOIN post_contents pc ON pc.post_id = p.id
		LEFT JOIN post_stats ps ON ps.post_id = p.id
		LEFT JOIN post_tags pt ON pt.post_id = p.id
		LEFT JOIN tags t ON t.id = pt.tag_id
		WHERE p.deleted_at IS NULL AND p.status = 'published'
	`
	listQuery += listArgs.query

	var total int
	if err := r.db.Raw(countQuery, countArgs.args...).Row().Scan(&total); err != nil {
		return nil, err
	}

	listQuery += `
		GROUP BY
			p.id, p.author_id, p.category_id, c.name, c.slug, p.title, p.slug, p.summary,
			pc.content_markdown, u.username, p.status, p.is_top, p.created_at, p.updated_at,
			p.published_at, ps.view_count, ps.like_count, ps.favorite_count, ps.comment_count
		ORDER BY tag_match_score DESC, text_match_score DESC, p.is_top DESC, p.published_at DESC, p.id DESC
		LIMIT ? OFFSET ?
	`
	listArgs.args = append(listArgs.args, pageSize, offset)

	rows, err := r.db.Raw(listQuery, listArgs.args...).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items, err := scanBlogRows(rows)
	if err != nil {
		return nil, err
	}

	return &model.BlogListResult{
		Items:      items,
		Page:       page,
		PageSize:   pageSize,
		Total:      total,
		Keyword:    keyword,
		CategoryID: query.CategoryID,
		Archive:    archive,
	}, nil
}

// AdminList 返回后台博客列表。
func (r *BlogRepository) AdminList(page, pageSize int, keyword, author, status string) (*model.BlogListResult, error) {
	keyword = strings.TrimSpace(keyword)
	author = strings.TrimSpace(author)
	status = strings.TrimSpace(status)
	offset := (page - 1) * pageSize
	likeKeyword := "%" + keyword + "%"
	likeAuthor := "%" + author + "%"

	countQuery := `
		SELECT COUNT(*)
		FROM posts p
		LEFT JOIN users u ON u.id = p.author_id
		LEFT JOIN post_contents pc ON pc.post_id = p.id
		WHERE p.deleted_at IS NULL
	`
	listQuery := `
		SELECT
			p.id,
			p.author_id,
			COALESCE(p.category_id, 0),
			COALESCE(c.name, ''),
			COALESCE(c.slug, ''),
			p.title,
			COALESCE(p.slug, ''),
			COALESCE(p.summary, ''),
			COALESCE(pc.content_markdown, ''),
			COALESCE(u.username, ''),
			p.status,
			p.is_top,
			p.created_at,
			p.updated_at,
			p.published_at,
			COALESCE(ps.view_count, 0),
			COALESCE(ps.like_count, 0),
			COALESCE(ps.favorite_count, 0),
			COALESCE(ps.comment_count, 0),
			COALESCE(GROUP_CONCAT(DISTINCT CONCAT(t.id, '::', t.name, '::', t.slug) ORDER BY t.name SEPARATOR '||'), '')
		FROM posts p
		LEFT JOIN users u ON u.id = p.author_id
		LEFT JOIN categories c ON c.id = p.category_id
		LEFT JOIN post_contents pc ON pc.post_id = p.id
		LEFT JOIN post_stats ps ON ps.post_id = p.id
		LEFT JOIN post_tags pt ON pt.post_id = p.id
		LEFT JOIN tags t ON t.id = pt.tag_id
		WHERE p.deleted_at IS NULL
	`

	var countArgs []any
	var listArgs []any
	if keyword != "" {
		filter := `
			AND (
				p.title LIKE ?
				OR COALESCE(p.summary, '') LIKE ?
				OR COALESCE(pc.content_text, pc.content_markdown, '') LIKE ?
				OR COALESCE(u.username, '') LIKE ?
			)
		`
		countQuery += filter
		listQuery += filter
		countArgs = append(countArgs, likeKeyword, likeKeyword, likeKeyword, likeKeyword)
		listArgs = append(listArgs, likeKeyword, likeKeyword, likeKeyword, likeKeyword)
	}
	if author != "" {
		filter := ` AND COALESCE(u.username, '') LIKE ?`
		countQuery += filter
		listQuery += filter
		countArgs = append(countArgs, likeAuthor)
		listArgs = append(listArgs, likeAuthor)
	}
	if status != "" {
		filter := ` AND p.status = ?`
		countQuery += filter
		listQuery += filter
		countArgs = append(countArgs, status)
		listArgs = append(listArgs, status)
	}

	var total int
	if err := r.db.Raw(countQuery, countArgs...).Row().Scan(&total); err != nil {
		return nil, err
	}

	listQuery += `
		GROUP BY
			p.id, p.author_id, p.category_id, c.name, c.slug, p.title, p.slug, p.summary,
			pc.content_markdown, u.username, p.status, p.is_top, p.created_at, p.updated_at,
			p.published_at, ps.view_count, ps.like_count, ps.favorite_count, ps.comment_count
		ORDER BY p.is_top DESC, p.created_at DESC, p.id DESC
		LIMIT ? OFFSET ?
	`
	listArgs = append(listArgs, pageSize, offset)

	rows, err := r.db.Raw(listQuery, listArgs...).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items, err := scanBlogRows(rows)
	if err != nil {
		return nil, err
	}

	return &model.BlogListResult{
		Items:    items,
		Page:     page,
		PageSize: pageSize,
		Total:    total,
		Keyword:  keyword,
	}, nil
}

// ListByAuthor 返回指定作者的博客列表。
func (r *BlogRepository) ListByAuthor(page, pageSize int, authorUsername, status string) (*model.BlogListResult, error) {
	authorUsername = strings.TrimSpace(authorUsername)
	status = strings.TrimSpace(status)
	offset := (page - 1) * pageSize

	query := `
		SELECT
			p.id,
			p.author_id,
			COALESCE(p.category_id, 0),
			COALESCE(c.name, ''),
			COALESCE(c.slug, ''),
			p.title,
			COALESCE(p.slug, ''),
			COALESCE(p.summary, ''),
			COALESCE(pc.content_markdown, ''),
			COALESCE(u.username, ''),
			p.status,
			p.is_top,
			p.created_at,
			p.updated_at,
			p.published_at,
			COALESCE(ps.view_count, 0),
			COALESCE(ps.like_count, 0),
			COALESCE(ps.favorite_count, 0),
			COALESCE(ps.comment_count, 0),
			COALESCE(GROUP_CONCAT(DISTINCT CONCAT(t.id, '::', t.name, '::', t.slug) ORDER BY t.name SEPARATOR '||'), '')
		FROM posts p
		LEFT JOIN users u ON u.id = p.author_id
		LEFT JOIN categories c ON c.id = p.category_id
		LEFT JOIN post_contents pc ON pc.post_id = p.id
		LEFT JOIN post_stats ps ON ps.post_id = p.id
		LEFT JOIN post_tags pt ON pt.post_id = p.id
		LEFT JOIN tags t ON t.id = pt.tag_id
		WHERE p.deleted_at IS NULL AND COALESCE(u.username, '') = ?
	`
	countQuery := `
		SELECT COUNT(*)
		FROM posts p
		LEFT JOIN users u ON u.id = p.author_id
		WHERE p.deleted_at IS NULL AND COALESCE(u.username, '') = ?
	`

	args := []any{authorUsername}
	countArgs := []any{authorUsername}
	if status != "" {
		query += ` AND p.status = ?`
		countQuery += ` AND p.status = ?`
		args = append(args, status)
		countArgs = append(countArgs, status)
	}

	var total int
	if err := r.db.Raw(countQuery, countArgs...).Row().Scan(&total); err != nil {
		return nil, err
	}

	query += `
		GROUP BY
			p.id, p.author_id, p.category_id, c.name, c.slug, p.title, p.slug, p.summary,
			pc.content_markdown, u.username, p.status, p.is_top, p.created_at, p.updated_at,
			p.published_at, ps.view_count, ps.like_count, ps.favorite_count, ps.comment_count
		ORDER BY p.updated_at DESC, p.id DESC
		LIMIT ? OFFSET ?
	`
	args = append(args, pageSize, offset)

	rows, err := r.db.Raw(query, args...).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items, err := scanBlogRows(rows)
	if err != nil {
		return nil, err
	}

	return &model.BlogListResult{
		Items:    items,
		Page:     page,
		PageSize: pageSize,
		Total:    total,
	}, nil
}

// ListFavoritesByUser 返回指定用户收藏的博客列表。
func (r *BlogRepository) ListFavoritesByUser(page, pageSize int, username string) (*model.BlogListResult, error) {
	username = strings.TrimSpace(username)
	offset := (page - 1) * pageSize

	countQuery := `
		SELECT COUNT(*)
		FROM post_favorites pf
		INNER JOIN users fav_u ON fav_u.id = pf.user_id
		INNER JOIN posts p ON p.id = pf.post_id
		WHERE fav_u.username = ? AND fav_u.deleted_at IS NULL
			AND p.deleted_at IS NULL AND p.status = 'published'
	`
	listQuery := `
		SELECT
			p.id,
			p.author_id,
			COALESCE(p.category_id, 0),
			COALESCE(c.name, ''),
			COALESCE(c.slug, ''),
			p.title,
			COALESCE(p.slug, ''),
			COALESCE(p.summary, ''),
			COALESCE(pc.content_markdown, ''),
			COALESCE(u.username, ''),
			p.status,
			p.is_top,
			p.created_at,
			p.updated_at,
			p.published_at,
			COALESCE(ps.view_count, 0),
			COALESCE(ps.like_count, 0),
			COALESCE(ps.favorite_count, 0),
			COALESCE(ps.comment_count, 0),
			COALESCE(GROUP_CONCAT(DISTINCT CONCAT(t.id, '::', t.name, '::', t.slug) ORDER BY t.name SEPARATOR '||'), '')
		FROM post_favorites pf
		INNER JOIN users fav_u ON fav_u.id = pf.user_id
		INNER JOIN posts p ON p.id = pf.post_id
		LEFT JOIN users u ON u.id = p.author_id
		LEFT JOIN categories c ON c.id = p.category_id
		LEFT JOIN post_contents pc ON pc.post_id = p.id
		LEFT JOIN post_stats ps ON ps.post_id = p.id
		LEFT JOIN post_tags pt ON pt.post_id = p.id
		LEFT JOIN tags t ON t.id = pt.tag_id
		WHERE fav_u.username = ? AND fav_u.deleted_at IS NULL
			AND p.deleted_at IS NULL AND p.status = 'published'
		GROUP BY
			p.id, p.author_id, p.category_id, c.name, c.slug, p.title, p.slug, p.summary,
			pc.content_markdown, u.username, p.status, p.is_top, p.created_at, p.updated_at,
			p.published_at, ps.view_count, ps.like_count, ps.favorite_count, ps.comment_count, pf.created_at
		ORDER BY pf.created_at DESC, p.id DESC
		LIMIT ? OFFSET ?
	`

	var total int
	if err := r.db.Raw(countQuery, username).Row().Scan(&total); err != nil {
		return nil, err
	}

	rows, err := r.db.Raw(listQuery, username, pageSize, offset).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items, err := scanBlogRows(rows)
	if err != nil {
		return nil, err
	}

	return &model.BlogListResult{
		Items:    items,
		Page:     page,
		PageSize: pageSize,
		Total:    total,
	}, nil
}

// GetByID 按文章 ID 读取博客详情。
func (r *BlogRepository) GetByID(blogID int64) (*model.Blog, error) {
	rows, err := r.db.Raw(`
		SELECT
			p.id,
			p.author_id,
			COALESCE(p.category_id, 0),
			COALESCE(c.name, ''),
			COALESCE(c.slug, ''),
			p.title,
			COALESCE(p.slug, ''),
			COALESCE(p.summary, ''),
			COALESCE(pc.content_markdown, ''),
			COALESCE(u.username, ''),
			p.status,
			p.is_top,
			p.created_at,
			p.updated_at,
			p.published_at,
			COALESCE(ps.view_count, 0),
			COALESCE(ps.like_count, 0),
			COALESCE(ps.favorite_count, 0),
			COALESCE(ps.comment_count, 0),
			COALESCE(GROUP_CONCAT(DISTINCT CONCAT(t.id, '::', t.name, '::', t.slug) ORDER BY t.name SEPARATOR '||'), '')
		FROM posts p
		LEFT JOIN users u ON u.id = p.author_id
		LEFT JOIN categories c ON c.id = p.category_id
		LEFT JOIN post_contents pc ON pc.post_id = p.id
		LEFT JOIN post_stats ps ON ps.post_id = p.id
		LEFT JOIN post_tags pt ON pt.post_id = p.id
		LEFT JOIN tags t ON t.id = pt.tag_id
		WHERE p.id = ? AND p.deleted_at IS NULL
		GROUP BY
			p.id, p.author_id, p.category_id, c.name, c.slug, p.title, p.slug, p.summary,
			pc.content_markdown, u.username, p.status, p.is_top, p.created_at, p.updated_at,
			p.published_at, ps.view_count, ps.like_count, ps.favorite_count, ps.comment_count
	`, blogID).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	blogs, err := scanBlogRows(rows)
	if err != nil {
		return nil, err
	}
	if len(blogs) == 0 {
		return nil, sql.ErrNoRows
	}
	return &blogs[0], nil
}

// Create 写入一篇新博客。
func (r *BlogRepository) Create(blog *model.Blog) error {
	tx := r.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer tx.Rollback()

	authorID, err := getUserIDByUsername(tx, blog.AuthorUsername)
	if err != nil {
		return err
	}

	if err := ensureCategoryExists(tx, blog.CategoryID); err != nil {
		return err
	}

	slug := blog.Slug
	if slug == "" {
		slug = fmt.Sprintf("post-%d", time.Now().UnixNano())
	}

	var publishedAt any
	if blog.Status == "published" {
		now := time.Now()
		publishedAt = now
		blog.PublishedAt = &now
	} else {
		publishedAt = nil
		blog.PublishedAt = nil
	}

	result := tx.Exec(`
		INSERT INTO posts (author_id, category_id, title, slug, summary, status, visibility, published_at, is_top)
		VALUES (?, NULLIF(?, 0), ?, ?, ?, ?, 'public', ?, ?)
	`, authorID, blog.CategoryID, blog.Title, slug, blog.Summary, blog.Status, publishedAt, blog.IsTop)
	if result.Error != nil {
		return result.Error
	}

	var postID int64
	if err := tx.Raw("SELECT LAST_INSERT_ID()").Row().Scan(&postID); err != nil {
		return err
	}
	blog.ID = postID
	blog.AuthorID = authorID

	if err := tx.Exec(`
		INSERT INTO post_contents (post_id, content_markdown, content_text, word_count)
		VALUES (?, ?, ?, ?)
	`, postID, blog.Content, buildPlainTextFromMarkdown(blog.Content), countWordsFromMarkdown(blog.Content)).Error; err != nil {
		return err
	}

	if err := tx.Exec("INSERT INTO post_stats (post_id) VALUES (?)", postID).Error; err != nil {
		return err
	}

	tags, err := replacePostTags(tx, postID, blog.Tags)
	if err != nil {
		return err
	}
	blog.Tags = tags

	if err := tx.Commit().Error; err != nil {
		return err
	}

	return r.populateCategory(blog)
}

// GetAuthorByID 按文章 ID 读取作者用户名。
func (r *BlogRepository) GetAuthorByID(blogID int64) (string, error) {
	var authorUsername string
	err := r.db.Raw(`
		SELECT COALESCE(u.username, '')
		FROM posts p
		LEFT JOIN users u ON u.id = p.author_id
		WHERE p.id = ? AND p.deleted_at IS NULL
	`, blogID).Row().Scan(&authorUsername)
	if err != nil {
		return "", err
	}
	return authorUsername, nil
}

// Update 更新博客内容和状态。
func (r *BlogRepository) Update(blog *model.Blog) error {
	tx := r.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer tx.Rollback()

	if err := ensureCategoryExists(tx, blog.CategoryID); err != nil {
		return err
	}

	var publishedAt any
	if blog.Status == "published" {
		if blog.PublishedAt != nil {
			publishedAt = *blog.PublishedAt
		} else {
			now := time.Now()
			publishedAt = now
			blog.PublishedAt = &now
		}
	} else {
		publishedAt = nil
	}

	if err := tx.Exec(`
		UPDATE posts
		SET category_id = NULLIF(?, 0), title = ?, summary = ?, status = ?, is_top = ?, published_at = ?, updated_at = NOW()
		WHERE id = ? AND deleted_at IS NULL
	`, blog.CategoryID, blog.Title, blog.Summary, blog.Status, blog.IsTop, publishedAt, blog.ID).Error; err != nil {
		return err
	}

	if err := tx.Exec(`
		UPDATE post_contents
		SET content_markdown = ?, content_text = ?, word_count = ?, updated_at = NOW()
		WHERE post_id = ?
	`, blog.Content, buildPlainTextFromMarkdown(blog.Content), countWordsFromMarkdown(blog.Content), blog.ID).Error; err != nil {
		return err
	}

	tags, err := replacePostTags(tx, blog.ID, blog.Tags)
	if err != nil {
		return err
	}
	blog.Tags = tags

	if err := tx.Commit().Error; err != nil {
		return err
	}

	return r.populateCategory(blog)
}

// Review 更新审核状态和置顶状态。
func (r *BlogRepository) Review(blogID int64, status string, isTop bool) error {
	var publishedAt any
	if status == "published" {
		publishedAt = time.Now()
	}

	result := r.db.Exec(`
		UPDATE posts
		SET status = ?, is_top = ?, published_at = ?, updated_at = NOW()
		WHERE id = ? AND deleted_at IS NULL
	`, status, isTop, publishedAt, blogID)
	if result.Error != nil {
		return result.Error
	}

	affected := result.RowsAffected
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// Delete 删除博客。
func (r *BlogRepository) Delete(blogID int64) error {
	return r.db.Exec("DELETE FROM posts WHERE id = ?", blogID).Error
}

// ListCategories 返回分类列表。
func (r *BlogRepository) ListCategories() ([]model.Category, error) {
	rows, err := r.db.Raw(`
		SELECT
			c.id,
			c.name,
			c.slug,
			c.status,
			COUNT(CASE WHEN p.deleted_at IS NULL AND p.status = 'published' THEN 1 END) AS post_count
		FROM categories c
		LEFT JOIN posts p ON p.category_id = c.id
		WHERE c.status = 'active'
		GROUP BY c.id, c.name, c.slug, c.sort_order
		ORDER BY c.sort_order ASC, c.id ASC
	`).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []model.Category
	for rows.Next() {
		var item model.Category
		if err := rows.Scan(&item.ID, &item.Name, &item.Slug, &item.Status, &item.PostCount); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

// ListCategoriesForManage 返回后台分类列表。
func (r *BlogRepository) ListCategoriesForManage() ([]model.Category, error) {
	rows, err := r.db.Raw(`
		SELECT
			c.id,
			c.name,
			c.slug,
			c.status,
			COUNT(CASE WHEN p.deleted_at IS NULL THEN 1 END) AS post_count
		FROM categories c
		LEFT JOIN posts p ON p.category_id = c.id
		GROUP BY c.id, c.name, c.slug, c.status, c.sort_order
		ORDER BY c.status = 'active' DESC, c.sort_order ASC, c.id ASC
	`).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []model.Category
	for rows.Next() {
		var item model.Category
		if err := rows.Scan(&item.ID, &item.Name, &item.Slug, &item.Status, &item.PostCount); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

// CreateCategory 创建分类。
func (r *BlogRepository) CreateCategory(category *model.Category) error {
	result := r.db.Exec(`
		INSERT INTO categories (name, slug, status)
		VALUES (?, ?, 'active')
	`, category.Name, category.Slug)
	if result.Error != nil {
		return result.Error
	}

	var categoryID int64
	if err := r.db.Raw("SELECT LAST_INSERT_ID()").Row().Scan(&categoryID); err != nil {
		return err
	}
	category.ID = categoryID
	return nil
}

// UpdateCategory 更新分类。
func (r *BlogRepository) UpdateCategory(category *model.Category) error {
	result := r.db.Exec(`
		UPDATE categories
		SET name = ?, slug = ?, status = 'active', updated_at = NOW()
		WHERE id = ?
	`, category.Name, category.Slug, category.ID)
	if result.Error != nil {
		return result.Error
	}

	affected := result.RowsAffected
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// HideCategory 隐藏分类。
func (r *BlogRepository) HideCategory(categoryID int64) error {
	result := r.db.Exec(`
		UPDATE categories
		SET status = 'hidden', updated_at = NOW()
		WHERE id = ?
	`, categoryID)
	if result.Error != nil {
		return result.Error
	}

	affected := result.RowsAffected
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// ListTags 返回标签列表。
func (r *BlogRepository) ListTags() ([]model.Tag, error) {
	rows, err := r.db.Raw(`
		SELECT
			t.id,
			t.name,
			t.slug,
			COUNT(CASE WHEN p.deleted_at IS NULL AND p.status = 'published' THEN 1 END) AS post_count
		FROM tags t
		LEFT JOIN post_tags pt ON pt.tag_id = t.id
		LEFT JOIN posts p ON p.id = pt.post_id
		GROUP BY t.id, t.name, t.slug
		ORDER BY post_count DESC, t.name ASC
	`).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []model.Tag
	for rows.Next() {
		var item model.Tag
		if err := rows.Scan(&item.ID, &item.Name, &item.Slug, &item.PostCount); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

// ListArchives 返回归档列表。
func (r *BlogRepository) ListArchives() ([]model.ArchiveItem, error) {
	rows, err := r.db.Raw(`
		SELECT
			DATE_FORMAT(p.published_at, '%Y-%m') AS archive,
			YEAR(p.published_at) AS year_num,
			MONTH(p.published_at) AS month_num,
			COUNT(*) AS total
		FROM posts p
		WHERE p.deleted_at IS NULL AND p.status = 'published' AND p.published_at IS NOT NULL
		GROUP BY archive, year_num, month_num
		ORDER BY archive DESC
	`).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []model.ArchiveItem
	for rows.Next() {
		var item model.ArchiveItem
		if err := rows.Scan(&item.Archive, &item.Year, &item.Month, &item.Count); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

// IncrementViewCount 增加阅读量。
func (r *BlogRepository) IncrementViewCount(blogID int64) error {
	return r.db.Exec(`
		UPDATE post_stats
		SET view_count = view_count + 1, updated_at = NOW()
		WHERE post_id = ?
	`, blogID).Error
}

// HasLiked 判断当前用户是否已点赞。
func (r *BlogRepository) HasLiked(blogID int64, username string) (bool, error) {
	return hasInteraction(r.db, "post_likes", blogID, username)
}

// HasFavorited 判断当前用户是否已收藏。
func (r *BlogRepository) HasFavorited(blogID int64, username string) (bool, error) {
	return hasInteraction(r.db, "post_favorites", blogID, username)
}

// ToggleLike 切换点赞状态并返回最新点赞数。
func (r *BlogRepository) ToggleLike(blogID int64, username string) (bool, int64, error) {
	return toggleInteraction(r.db, "post_likes", "like_count", blogID, username)
}

// ToggleFavorite 切换收藏状态并返回最新收藏数。
func (r *BlogRepository) ToggleFavorite(blogID int64, username string) (bool, int64, error) {
	return toggleInteraction(r.db, "post_favorites", "favorite_count", blogID, username)
}

func (r *BlogRepository) populateCategory(blog *model.Blog) error {
	if blog.CategoryID == 0 {
		blog.CategoryName = ""
		blog.CategorySlug = ""
		return nil
	}

	var name string
	var slug string
	if err := r.db.Raw(`
		SELECT name, slug
		FROM categories
		WHERE id = ?
	`, blog.CategoryID).Row().Scan(&name, &slug); err != nil {
		return err
	}
	blog.CategoryName = name
	blog.CategorySlug = slug
	return nil
}

type listFilter struct {
	query string
	args  []any
}

func buildBlogListFilters(keyword string, categoryID int64, archive string) (listFilter, listFilter) {
	likeKeyword := "%" + keyword + "%"
	count := listFilter{}
	list := listFilter{}

	if keyword != "" {
		filter := `
			AND (
				p.title LIKE ?
				OR COALESCE(p.summary, '') LIKE ?
				OR COALESCE(pc.content_text, pc.content_markdown, '') LIKE ?
				OR COALESCE(u.username, '') LIKE ?
				OR EXISTS (
					SELECT 1
					FROM post_tags pt2
					INNER JOIN tags t2 ON t2.id = pt2.tag_id
					WHERE pt2.post_id = p.id
						AND (COALESCE(t2.name, '') LIKE ? OR COALESCE(t2.slug, '') LIKE ?)
				)
			)
		`
		count.query += filter
		list.query += filter
		for _, target := range []*listFilter{&count, &list} {
			target.args = append(target.args, likeKeyword, likeKeyword, likeKeyword, likeKeyword, likeKeyword, likeKeyword)
		}
	}

	if categoryID > 0 {
		filter := ` AND p.category_id = ?`
		count.query += filter
		list.query += filter
		count.args = append(count.args, categoryID)
		list.args = append(list.args, categoryID)
	}

	if archive != "" {
		filter := ` AND DATE_FORMAT(p.published_at, '%Y-%m') = ?`
		count.query += filter
		list.query += filter
		count.args = append(count.args, archive)
		list.args = append(list.args, archive)
	}

	return count, list
}

func scanBlogRows(rows *sql.Rows) ([]model.Blog, error) {
	var blogs []model.Blog
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	hasMatchScores := len(columns) >= 22

	for rows.Next() {
		var blog model.Blog
		var publishedAt sql.NullTime
		var tagTokens string
		var tagMatchScore int
		var textMatchScore int

		dest := []any{
			&blog.ID,
			&blog.AuthorID,
			&blog.CategoryID,
			&blog.CategoryName,
			&blog.CategorySlug,
			&blog.Title,
			&blog.Slug,
			&blog.Summary,
			&blog.Content,
			&blog.AuthorUsername,
			&blog.Status,
			&blog.IsTop,
			&blog.CreatedAt,
			&blog.UpdatedAt,
			&publishedAt,
			&blog.Stats.ViewCount,
			&blog.Stats.LikeCount,
			&blog.Stats.FavoriteCount,
			&blog.Stats.CommentCount,
			&tagTokens,
		}
		if hasMatchScores {
			dest = append(dest, &tagMatchScore, &textMatchScore)
		}

		if err := rows.Scan(dest...); err != nil {
			return nil, err
		}

		if publishedAt.Valid {
			blog.PublishedAt = &publishedAt.Time
		}
		blog.Tags = parseTagTokens(tagTokens)
		blogs = append(blogs, blog)
	}

	return blogs, rows.Err()
}

func parseTagTokens(raw string) []model.Tag {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}

	parts := strings.Split(raw, "||")
	items := make([]model.Tag, 0, len(parts))
	for _, part := range parts {
		fields := strings.Split(part, "::")
		if len(fields) != 3 {
			continue
		}

		tagID, err := strconv.ParseInt(fields[0], 10, 64)
		if err != nil {
			continue
		}

		items = append(items, model.Tag{
			ID:   tagID,
			Name: fields[1],
			Slug: fields[2],
		})
	}

	return items
}

func getUserIDByUsername(tx *gorm.DB, username string) (int64, error) {
	var userID int64
	if err := tx.Raw(`
		SELECT id
		FROM users
		WHERE username = ? AND deleted_at IS NULL
	`, username).Row().Scan(&userID); err != nil {
		return 0, err
	}
	return userID, nil
}

func ensureCategoryExists(tx *gorm.DB, categoryID int64) error {
	if categoryID == 0 {
		return nil
	}

	var exists int
	if err := tx.Raw(`
		SELECT 1
		FROM categories
		WHERE id = ? AND status = 'active'
	`, categoryID).Row().Scan(&exists); err != nil {
		if err == sql.ErrNoRows {
			return errorsNew("category not found")
		}
		return err
	}
	return nil
}

func replacePostTags(tx *gorm.DB, postID int64, tags []model.Tag) ([]model.Tag, error) {
	if err := tx.Exec(`DELETE FROM post_tags WHERE post_id = ?`, postID).Error; err != nil {
		return nil, err
	}

	if len(tags) == 0 {
		return nil, nil
	}

	result := make([]model.Tag, 0, len(tags))
	for _, tag := range tags {
		tagName := strings.TrimSpace(tag.Name)
		tagSlug := strings.TrimSpace(tag.Slug)
		if tagName == "" || tagSlug == "" {
			continue
		}

		var tagID int64
		err := tx.Raw(`
			SELECT id
			FROM tags
			WHERE slug = ?
		`, tagSlug).Row().Scan(&tagID)
		if err != nil {
			if err == sql.ErrNoRows {
				if err := tx.Exec(`INSERT INTO tags (name, slug) VALUES (?, ?)`, tagName, tagSlug).Error; err != nil {
					return nil, err
				}
				if err := tx.Raw("SELECT LAST_INSERT_ID()").Row().Scan(&tagID); err != nil {
					return nil, err
				}
			} else {
				return nil, err
			}
		}

		if err := tx.Exec(`
			INSERT INTO post_tags (post_id, tag_id)
			VALUES (?, ?)
		`, postID, tagID).Error; err != nil {
			return nil, err
		}

		result = append(result, model.Tag{
			ID:   tagID,
			Name: tagName,
			Slug: tagSlug,
		})
	}

	return result, nil
}

func hasInteraction(db *gorm.DB, table string, blogID int64, username string) (bool, error) {
	username = strings.TrimSpace(username)
	if username == "" {
		return false, nil
	}

	var exists int
	err := db.Raw(fmt.Sprintf(`
		SELECT 1
		FROM %s pi
		INNER JOIN users u ON u.id = pi.user_id
		WHERE pi.post_id = ? AND u.username = ? AND u.deleted_at IS NULL
	`, table), blogID, username).Row().Scan(&exists)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func toggleInteraction(db *gorm.DB, table, statColumn string, blogID int64, username string) (bool, int64, error) {
	tx := db.Begin()
	if tx.Error != nil {
		return false, 0, tx.Error
	}
	defer tx.Rollback()

	userID, err := getUserIDByUsername(tx, username)
	if err != nil {
		return false, 0, err
	}

	var exists int
	err = tx.Raw(fmt.Sprintf(`
		SELECT 1
		FROM %s
		WHERE post_id = ? AND user_id = ?
	`, table), blogID, userID).Row().Scan(&exists)
	active := false

	switch err {
	case nil:
		if err := tx.Exec(fmt.Sprintf(`
			DELETE FROM %s
			WHERE post_id = ? AND user_id = ?
		`, table), blogID, userID).Error; err != nil {
			return false, 0, err
		}

		if err := tx.Exec(fmt.Sprintf(`
			UPDATE post_stats
			SET %s = CASE WHEN %s > 0 THEN %s - 1 ELSE 0 END, updated_at = NOW()
			WHERE post_id = ?
		`, statColumn, statColumn, statColumn), blogID).Error; err != nil {
			return false, 0, err
		}
	case sql.ErrNoRows:
		if err := tx.Exec(fmt.Sprintf(`
			INSERT INTO %s (post_id, user_id)
			VALUES (?, ?)
		`, table), blogID, userID).Error; err != nil {
			return false, 0, err
		}

		if err := tx.Exec(fmt.Sprintf(`
			UPDATE post_stats
			SET %s = %s + 1, updated_at = NOW()
			WHERE post_id = ?
		`, statColumn, statColumn), blogID).Error; err != nil {
			return false, 0, err
		}
		active = true
	default:
		return false, 0, err
	}

	var count int64
	if err := tx.Raw(fmt.Sprintf(`
		SELECT %s
		FROM post_stats
		WHERE post_id = ?
	`, statColumn), blogID).Row().Scan(&count); err != nil {
		return false, 0, err
	}

	if err := tx.Commit().Error; err != nil {
		return false, 0, err
	}

	return active, count, nil
}

var markdownCleanupReplacer = strings.NewReplacer(
	"\r", " ",
	"\n", " ",
	"\t", " ",
	"#", " ",
	"*", " ",
	"_", " ",
	"`", " ",
	">", " ",
	"-", " ",
	"|", " ",
	"[", " ",
	"]", " ",
	"(", " ",
	")", " ",
	"!", " ",
)

var multiSpaceRegexp = regexp.MustCompile(`\s+`)

// buildPlainTextFromMarkdown 去掉 Markdown 标记并压缩空白。
func buildPlainTextFromMarkdown(content string) string {
	plain := markdownCleanupReplacer.Replace(content)
	plain = multiSpaceRegexp.ReplaceAllString(plain, " ")
	return strings.TrimSpace(plain)
}

// countWordsFromMarkdown 统计 Markdown 纯文本长度。
func countWordsFromMarkdown(content string) int {
	return len([]rune(buildPlainTextFromMarkdown(content)))
}

func errorsNew(message string) error {
	return fmt.Errorf("%s", message)
}
