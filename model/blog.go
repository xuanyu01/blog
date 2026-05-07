/*
定义博客、分类、标签、归档和互动相关模型。
*/
package model

import "time"

// Blog 表示一篇博客文章。
type Blog struct {
	ID             int64      `json:"id"`
	AuthorID       int64      `json:"authorId"`
	CategoryID     int64      `json:"categoryId"`
	CategoryName   string     `json:"categoryName"`
	CategorySlug   string     `json:"categorySlug"`
	Title          string     `json:"title"`
	Slug           string     `json:"slug"`
	Summary        string     `json:"summary"`
	Content        string     `json:"content"`
	AuthorUsername string     `json:"authorUsername"`
	Status         string     `json:"status"`
	IsTop          bool       `json:"isTop"`
	CreatedAt      time.Time  `json:"createdAt"`
	UpdatedAt      time.Time  `json:"updatedAt"`
	PublishedAt    *time.Time `json:"publishedAt"`
	Tags           []Tag      `json:"tags"`
	Stats          BlogStats  `json:"stats"`
	Liked          bool       `json:"liked"`
	Favorited      bool       `json:"favorited"`
}

// BlogStats 表示文章互动统计。
type BlogStats struct {
	ViewCount     int64 `json:"viewCount"`
	LikeCount     int64 `json:"likeCount"`
	FavoriteCount int64 `json:"favoriteCount"`
	CommentCount  int64 `json:"commentCount"`
}

// BlogListQuery 表示博客列表筛选条件。
type BlogListQuery struct {
	Keyword    string
	CategoryID int64
	Tag        string
	Archive    string
}

// BlogCreateInput 表示创建博客所需字段。
type BlogCreateInput struct {
	Title          string
	Content        string
	Status         string
	IsTop          bool
	CategoryID     int64
	Tags           []string
	AuthorUsername string
	Permission     string
}

// BlogUpdateInput 表示更新博客所需字段。
type BlogUpdateInput struct {
	BlogID      int64
	Title       string
	Content     string
	Status      string
	IsTop       bool
	CategoryID  int64
	Tags        []string
	CurrentUser string
	CurrentPerm string
}

// BlogInteraction 表示点赞收藏结果。
type BlogInteraction struct {
	Active        bool  `json:"active"`
	LikeCount     int64 `json:"likeCount"`
	FavoriteCount int64 `json:"favoriteCount"`
}

// BlogListResult 表示分页后的博客列表结果。
type BlogListResult struct {
	Items      []Blog `json:"items"`
	Page       int    `json:"page"`
	PageSize   int    `json:"pageSize"`
	Total      int    `json:"total"`
	Keyword    string `json:"keyword"`
	CategoryID int64  `json:"categoryId"`
	Tag        string `json:"tag"`
	Archive    string `json:"archive"`
}

// Category 表示文章分类。
type Category struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Slug      string `json:"slug"`
	Status    string `json:"status"`
	PostCount int64  `json:"postCount"`
}

// Tag 表示文章标签。
type Tag struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Slug      string `json:"slug"`
	PostCount int64  `json:"postCount"`
}

// ArchiveItem 表示归档分组。
type ArchiveItem struct {
	Archive string `json:"archive"`
	Year    int    `json:"year"`
	Month   int    `json:"month"`
	Count   int64  `json:"count"`
}
