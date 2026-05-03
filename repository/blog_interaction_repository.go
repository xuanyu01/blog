/*
blog_interaction_repository.go 。。。。。Ķ。。。。。。。。޺。。ղصĽ。。。。。。ݴ。。。。
*/
package repository

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

