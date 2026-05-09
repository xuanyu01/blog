/*
负责博客写入、更新、删除以及分类管理逻辑。
*/
package repository

import (
	"blog/model"
	"database/sql"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// Create 写入一篇新博客。
func (r *BlogRepository) Create(blog *model.Blog) error {
	var postID int64
	err := r.db.Transaction(func(tx *gorm.DB) error {
		// 先解析作者 ID，避免后续写入文章时丢失作者关系。
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

		return nil
	})
	if err != nil {
		return err
	}

	return r.populateCategory(blog)
}

// Update 更新博客内容和状态。
func (r *BlogRepository) Update(blog *model.Blog) error {
	err := r.db.Transaction(func(tx *gorm.DB) error {
		// 更新主表和正文表后，重新同步标签关系。
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

		return nil
	})
	if err != nil {
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

	if result.RowsAffected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// Delete 软删除博客，保留文章正文、标签、统计、评论和互动数据。
func (r *BlogRepository) Delete(blogID int64) error {
	result := r.db.Exec(`
		UPDATE posts
		SET deleted_at = NOW(), updated_at = NOW(), status = 'hidden'
		WHERE id = ? AND deleted_at IS NULL
	`, blogID)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return sql.ErrNoRows
	}
	return nil
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

	if result.RowsAffected == 0 {
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

	if result.RowsAffected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// populateCategory 回填博客上的分类名称和分类别名。
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
