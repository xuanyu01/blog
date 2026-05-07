/*
负责博客、分类、标签、归档和收藏列表的查询逻辑。
*/
package repository

import (
	"blog/model"
	"database/sql"
	"strings"
)

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
			CASE WHEN u.id IS NULL OR u.status = 'deleted' THEN '用户已注销' ELSE u.username END,
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
			pc.content_markdown, u.id, u.username, u.status, p.status, p.is_top, p.created_at, p.updated_at,
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
			CASE WHEN u.id IS NULL OR u.status = 'deleted' THEN '用户已注销' ELSE u.username END,
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
				OR CASE WHEN u.id IS NULL OR u.status = 'deleted' THEN '用户已注销' ELSE u.username END LIKE ?
			)
		`
		countQuery += filter
		listQuery += filter
		countArgs = append(countArgs, likeKeyword, likeKeyword, likeKeyword, likeKeyword)
		listArgs = append(listArgs, likeKeyword, likeKeyword, likeKeyword, likeKeyword)
	}
	if author != "" {
		filter := ` AND CASE WHEN u.id IS NULL OR u.status = 'deleted' THEN '用户已注销' ELSE u.username END LIKE ?`
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
			pc.content_markdown, u.id, u.username, u.status, p.status, p.is_top, p.created_at, p.updated_at,
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
			CASE WHEN u.id IS NULL OR u.status = 'deleted' THEN '用户已注销' ELSE u.username END,
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
		WHERE p.deleted_at IS NULL AND CASE WHEN u.id IS NULL OR u.status = 'deleted' THEN '用户已注销' ELSE u.username END = ?
	`
	countQuery := `
		SELECT COUNT(*)
		FROM posts p
		LEFT JOIN users u ON u.id = p.author_id
		WHERE p.deleted_at IS NULL AND CASE WHEN u.id IS NULL OR u.status = 'deleted' THEN '用户已注销' ELSE u.username END = ?
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
			pc.content_markdown, u.id, u.username, u.status, p.status, p.is_top, p.created_at, p.updated_at,
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
		WHERE fav_u.username = ? AND fav_u.deleted_at IS NULL AND fav_u.status <> 'deleted'
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
			CASE WHEN u.id IS NULL OR u.status = 'deleted' THEN '用户已注销' ELSE u.username END,
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
		WHERE fav_u.username = ? AND fav_u.deleted_at IS NULL AND fav_u.status <> 'deleted'
			AND p.deleted_at IS NULL AND p.status = 'published'
		GROUP BY
			p.id, p.author_id, p.category_id, c.name, c.slug, p.title, p.slug, p.summary,
			pc.content_markdown, u.id, u.username, u.status, p.status, p.is_top, p.created_at, p.updated_at,
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

// ListLikesByUser 返回指定用户点赞过的博客列表。
func (r *BlogRepository) ListLikesByUser(page, pageSize int, username string) (*model.BlogListResult, error) {
	username = strings.TrimSpace(username)
	offset := (page - 1) * pageSize

	countQuery := `
		SELECT COUNT(*)
		FROM post_likes pl
		INNER JOIN users liked_u ON liked_u.id = pl.user_id
		INNER JOIN posts p ON p.id = pl.post_id
		WHERE liked_u.username = ? AND liked_u.deleted_at IS NULL AND liked_u.status <> 'deleted'
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
			CASE WHEN u.id IS NULL OR u.status = 'deleted' THEN '用户已注销' ELSE u.username END,
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
		FROM post_likes pl
		INNER JOIN users liked_u ON liked_u.id = pl.user_id
		INNER JOIN posts p ON p.id = pl.post_id
		LEFT JOIN users u ON u.id = p.author_id
		LEFT JOIN categories c ON c.id = p.category_id
		LEFT JOIN post_contents pc ON pc.post_id = p.id
		LEFT JOIN post_stats ps ON ps.post_id = p.id
		LEFT JOIN post_tags pt ON pt.post_id = p.id
		LEFT JOIN tags t ON t.id = pt.tag_id
		WHERE liked_u.username = ? AND liked_u.deleted_at IS NULL AND liked_u.status <> 'deleted'
			AND p.deleted_at IS NULL AND p.status = 'published'
		GROUP BY
			p.id, p.author_id, p.category_id, c.name, c.slug, p.title, p.slug, p.summary,
			pc.content_markdown, u.id, u.username, u.status, p.status, p.is_top, p.created_at, p.updated_at,
			p.published_at, ps.view_count, ps.like_count, ps.favorite_count, ps.comment_count, pl.created_at
		ORDER BY pl.created_at DESC, p.id DESC
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
			CASE WHEN u.id IS NULL OR u.status = 'deleted' THEN '用户已注销' ELSE u.username END,
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
			pc.content_markdown, u.id, u.username, u.status, p.status, p.is_top, p.created_at, p.updated_at,
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

// GetAuthorByID 按文章 ID 读取作者用户名。
func (r *BlogRepository) GetAuthorByID(blogID int64) (string, error) {
	var authorUsername string
	err := r.db.Raw(`
		SELECT CASE WHEN u.id IS NULL OR u.status = 'deleted' THEN '用户已注销' ELSE u.username END
		FROM posts p
		LEFT JOIN users u ON u.id = p.author_id
		WHERE p.id = ? AND p.deleted_at IS NULL
	`, blogID).Row().Scan(&authorUsername)
	if err != nil {
		return "", err
	}
	return authorUsername, nil
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
