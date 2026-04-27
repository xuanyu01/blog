/*
实现评论数据访问逻辑。
*/
package repository

import (
	"blog/model"
	"database/sql"
)

// CommentRepository 负责评论读写。
type CommentRepository struct {
	db *sql.DB
}

// NewCommentRepository 创建评论仓储。
func NewCommentRepository(db *sql.DB) *CommentRepository {
	return &CommentRepository{db: db}
}

// ListByPostID 返回指定博客下的一级评论。
func (r *CommentRepository) ListByPostID(postID int64) ([]model.Comment, error) {
	rows, err := r.db.Query(`
		SELECT
			c.id,
			c.post_id,
			c.user_id,
			COALESCE(u.username, ''),
			COALESCE(NULLIF(u.display_name, ''), u.username, ''),
			c.content,
			c.created_at
		FROM comments c
		INNER JOIN users u ON u.id = c.user_id
		WHERE c.post_id = ?
			AND c.parent_id IS NULL
			AND c.deleted_at IS NULL
			AND c.status = 'published'
		ORDER BY c.created_at ASC, c.id ASC
	`, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []model.Comment
	for rows.Next() {
		var comment model.Comment
		if err := rows.Scan(
			&comment.ID,
			&comment.PostID,
			&comment.UserID,
			&comment.Username,
			&comment.DisplayName,
			&comment.Content,
			&comment.CreatedAt,
		); err != nil {
			return nil, err
		}
		comments = append(comments, comment)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return comments, nil
}

// Create 写入一条一级评论并返回详情。
func (r *CommentRepository) Create(postID int64, username, content string) (*model.Comment, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var userID int64
	var displayName string
	if err := tx.QueryRow(`
		SELECT id, COALESCE(NULLIF(display_name, ''), username, '')
		FROM users
		WHERE username = ? AND deleted_at IS NULL
	`, username).Scan(&userID, &displayName); err != nil {
		return nil, err
	}

	result, err := tx.Exec(`
		INSERT INTO comments (post_id, user_id, content, status)
		VALUES (?, ?, ?, 'published')
	`, postID, userID, content)
	if err != nil {
		return nil, err
	}

	commentID, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	comment := &model.Comment{
		ID:          commentID,
		PostID:      postID,
		UserID:      userID,
		Username:    username,
		DisplayName: displayName,
		Content:     content,
	}
	if err := tx.QueryRow(`
		SELECT created_at
		FROM comments
		WHERE id = ?
	`, commentID).Scan(&comment.CreatedAt); err != nil {
		return nil, err
	}

	if _, err := tx.Exec(`
		UPDATE post_stats
		SET comment_count = comment_count + 1, updated_at = NOW()
		WHERE post_id = ?
	`, postID); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return comment, nil
}

// GetAuthorByID 返回评论作者用户名。
func (r *CommentRepository) GetAuthorByID(commentID int64) (string, error) {
	var username string
	err := r.db.QueryRow(`
		SELECT COALESCE(u.username, '')
		FROM comments c
		INNER JOIN users u ON u.id = c.user_id
		WHERE c.id = ? AND c.deleted_at IS NULL
	`, commentID).Scan(&username)
	if err != nil {
		return "", err
	}
	return username, nil
}

// Delete 软删除评论并同步文章评论数。
func (r *CommentRepository) Delete(commentID int64) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var postID int64
	if err := tx.QueryRow(`
		SELECT post_id
		FROM comments
		WHERE id = ? AND deleted_at IS NULL
	`, commentID).Scan(&postID); err != nil {
		return err
	}

	result, err := tx.Exec(`
		UPDATE comments
		SET deleted_at = NOW(), updated_at = NOW(), status = 'hidden'
		WHERE id = ? AND deleted_at IS NULL
	`, commentID)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return sql.ErrNoRows
	}

	if _, err := tx.Exec(`
		UPDATE post_stats
		SET comment_count = CASE WHEN comment_count > 0 THEN comment_count - 1 ELSE 0 END, updated_at = NOW()
		WHERE post_id = ?
	`, postID); err != nil {
		return err
	}

	return tx.Commit()
}
