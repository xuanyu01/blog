/*
这个文件定义博客文章的领域模型
*/
package model

import "time"

// Blog 表示一篇博客的基础数据
type Blog struct {
	ID             int64
	AuthorID       int64
	Title          string
	Slug           string
	Summary        string
	Content        string
	AuthorUsername string
	CreatedAt      time.Time
}

// BlogListResult 表示分页后的博客列表结果
type BlogListResult struct {
	Items    []Blog
	Page     int
	PageSize int
	Total    int
	Keyword  string
}
