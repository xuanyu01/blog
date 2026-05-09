/*
负责评论数据的查询和写入逻辑。
*/
package repository

import (
	"blog/model"
	"database/sql"
	"errors"

	"gorm.io/gorm"
)

// CommentRepository 负责评论读写。
type CommentRepository struct {
	db *gorm.DB
}

// NewCommentRepository 创建评论仓储。
func NewCommentRepository(db *gorm.DB) *CommentRepository {
	return &CommentRepository{db: db}
}

// ListByPostID 返回指定博客下可见的评论树。
func (r *CommentRepository) ListByPostID(postID int64) ([]model.Comment, error) {
	rows, err := r.db.Raw(`
		SELECT
			c.id,
			c.post_id,
			c.user_id,
			c.parent_id,
			c.root_id,
			CASE WHEN u.id IS NULL OR u.status = 'deleted' THEN '用户已注销' ELSE COALESCE(u.username, '') END,
			CASE WHEN u.id IS NULL OR u.status = 'deleted' THEN '用户已注销' ELSE COALESCE(NULLIF(u.display_name, ''), u.username, '') END,
			CASE WHEN parent_u.id IS NULL OR parent_u.status = 'deleted' THEN '用户已注销' ELSE COALESCE(parent_u.username, '') END,
			c.content,
			c.created_at
		FROM comments c
		LEFT JOIN users u ON u.id = c.user_id
		LEFT JOIN comments parent_c ON parent_c.id = c.parent_id
		LEFT JOIN users parent_u ON parent_u.id = parent_c.user_id
		WHERE c.post_id = ?
			AND c.deleted_at IS NULL
			AND c.status = 'published'
		ORDER BY COALESCE(c.root_id, c.id) ASC, c.parent_id IS NOT NULL ASC, c.created_at ASC, c.id ASC
	`, postID).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	comments := make([]model.Comment, 0)
	for rows.Next() {
		var comment model.Comment
		var parentID sql.NullInt64
		var rootID sql.NullInt64
		if err := rows.Scan(
			&comment.ID,
			&comment.PostID,
			&comment.UserID,
			&parentID,
			&rootID,
			&comment.Username,
			&comment.DisplayName,
			&comment.ReplyToUsername,
			&comment.Content,
			&comment.CreatedAt,
		); err != nil {
			return nil, err
		}
		if parentID.Valid {
			value := parentID.Int64
			comment.ParentID = &value
		}
		if rootID.Valid {
			value := rootID.Int64
			comment.RootID = &value
		}
		comments = append(comments, comment)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return buildCommentTree(comments), nil
}

func buildCommentTree(comments []model.Comment) []model.Comment {
	roots := make([]model.Comment, 0)
	rootIndex := make(map[int64]int)

	for _, comment := range comments {
		if comment.ParentID == nil {
			comment.Replies = nil
			rootIndex[comment.ID] = len(roots)
			roots = append(roots, comment)
			continue
		}

		rootID := *comment.ParentID
		if comment.RootID != nil {
			rootID = *comment.RootID
		}
		if index, ok := rootIndex[rootID]; ok {
			roots[index].Replies = append(roots[index].Replies, comment)
		}
	}

	return roots
}

// Create 写入一级评论或回复并返回新评论
func (r *CommentRepository) Create(postID int64, parentID int64, username, content string) (*model.Comment, error) {
	var comment *model.Comment
	// 使用事务确保评论创建和文章评论数更新的原子性
	err := r.db.Transaction(func(tx *gorm.DB) error {
		var userID int64
		var displayName string
		// 通过用户名查询用户ID和显示名称，确保用户存在且未被删除
		if err := tx.Raw(`
			SELECT id, COALESCE(NULLIF(display_name, ''), username, '')
			FROM users
			WHERE username = ? AND deleted_at IS NULL
		`, username).Row().Scan(&userID, &displayName); err != nil {
			return err
		}

		// 验证博客存在且用户有权限评论
		var insertParentID any
		var insertRootID any
		var replyToUsername string
		if parentID > 0 {
			var parentPostID int64
			var parentRootID sql.NullInt64
			if err := tx.Raw(`
				SELECT c.post_id, c.root_id, CASE WHEN u.id IS NULL OR u.status = 'deleted' THEN '用户已注销' ELSE COALESCE(u.username, '') END
				FROM comments c
				LEFT JOIN users u ON u.id = c.user_id
				WHERE c.id = ? AND c.deleted_at IS NULL AND c.status = 'published'
			`, parentID).Row().Scan(&parentPostID, &parentRootID, &replyToUsername); err != nil {
				return err
			}
			if parentPostID != postID {
				return errors.New("parent comment does not belong to this blog")
			}

			insertParentID = parentID
			if parentRootID.Valid {
				insertRootID = parentRootID.Int64
			} else {
				insertRootID = parentID
			}
		}

		result := tx.Exec(`
			INSERT INTO comments (post_id, user_id, parent_id, root_id, content, status)
			VALUES (?, ?, ?, ?, ?, 'published')
		`, postID, userID, insertParentID, insertRootID, content)
		if result.Error != nil {
			return result.Error
		}

		var commentID int64
		if err := tx.Raw("SELECT LAST_INSERT_ID()").Row().Scan(&commentID); err != nil {
			return err
		}

		comment = &model.Comment{
			ID:              commentID,
			PostID:          postID,
			UserID:          userID,
			Username:        username,
			DisplayName:     displayName,
			ReplyToUsername: replyToUsername,
			Content:         content,
		}
		if parentID > 0 {
			comment.ParentID = &parentID
			if rootID, ok := insertRootID.(int64); ok {
				comment.RootID = &rootID
			}
		}
		if err := tx.Raw(`
			SELECT created_at
			FROM comments
			WHERE id = ?
		`, commentID).Row().Scan(&comment.CreatedAt); err != nil {
			return err
		}

		if err := tx.Exec(`
			UPDATE post_stats
			SET comment_count = comment_count + 1, updated_at = NOW()
			WHERE post_id = ?
		`, postID).Error; err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return comment, nil
}

// GetAuthorByID 返回评论作者用户名。
func (r *CommentRepository) GetAuthorByID(commentID int64) (string, error) {
	var username string
	err := r.db.Raw(`
		SELECT CASE WHEN u.id IS NULL OR u.status = 'deleted' THEN '用户已注销' ELSE COALESCE(u.username, '') END
		FROM comments c
		LEFT JOIN users u ON u.id = c.user_id
		WHERE c.id = ? AND c.deleted_at IS NULL
	`, commentID).Row().Scan(&username)
	if err != nil {
		return "", err
	}
	return username, nil
}

// Delete 软删除评论并同步文章评论数。
func (r *CommentRepository) Delete(commentID int64) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var postID int64
		if err := tx.Raw(`
			SELECT post_id
			FROM comments
			WHERE id = ? AND deleted_at IS NULL
		`, commentID).Row().Scan(&postID); err != nil {
			return err
		}

		result := tx.Exec(`
			UPDATE comments
			SET deleted_at = NOW(), updated_at = NOW(), status = 'hidden'
			WHERE id = ? AND deleted_at IS NULL
		`, commentID)
		if result.Error != nil {
			return result.Error
		}

		if result.RowsAffected == 0 {
			return sql.ErrNoRows
		}

		if err := tx.Exec(`
			UPDATE post_stats
			SET comment_count = CASE WHEN comment_count > 0 THEN comment_count - 1 ELSE 0 END, updated_at = NOW()
			WHERE post_id = ?
		`, postID).Error; err != nil {
			return err
		}

		return nil
	})
}
