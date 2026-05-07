/*
定义博客仓储结构以及查询、标签、Markdown 处理等公共辅助逻辑。
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
)

// BlogRepository 负责博客相关数据访问。
type BlogRepository struct {
	db *gorm.DB
}

// NewBlogRepository 创建博客仓储。
func NewBlogRepository(db *gorm.DB) *BlogRepository {
	return &BlogRepository{db: db}
}

type listFilter struct {
	query string
	args  []any
}

// buildBlogListFilters 构造博客列表查询和计数共用的筛选条件。
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
				OR CASE WHEN u.id IS NULL OR u.status = 'deleted' THEN '用户已注销' ELSE u.username END LIKE ?
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

// scanBlogRows 把查询结果逐行扫描为博客列表。
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

		// 兼容带匹配分数字段和不带匹配分数字段的查询结果。
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

// parseTagTokens 把聚合后的标签字符串还原为标签切片。
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

// getUserIDByUsername 按用户名读取用户主键。
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

// ensureCategoryExists 校验分类存在且处于可用状态。
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

// replacePostTags 重建文章与标签之间的关联关系。
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

		// 标签不存在时先补建，再建立文章和标签的关联。
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

// hasInteraction 判断用户是否已经对文章执行过某种互动。
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
		WHERE pi.post_id = ? AND u.username = ? AND u.deleted_at IS NULL AND u.status <> 'deleted'
	`, table), blogID, username).Row().Scan(&exists)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// toggleInteraction 切换点赞或收藏状态并同步更新统计值。
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

	// 先切换互动记录，再回读最新统计值。
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

// buildPlainTextFromMarkdown 生成用于摘要和检索的纯文本内容。
func buildPlainTextFromMarkdown(content string) string {
	plain := markdownCleanupReplacer.Replace(content)
	plain = multiSpaceRegexp.ReplaceAllString(plain, " ")
	return strings.TrimSpace(plain)
}

// countWordsFromMarkdown 统计 Markdown 内容的字符数。
func countWordsFromMarkdown(content string) int {
	return len([]rune(buildPlainTextFromMarkdown(content)))
}

// errorsNew 创建简单的错误对象。
func errorsNew(message string) error {
	return fmt.Errorf("%s", message)
}
