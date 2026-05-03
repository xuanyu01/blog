/*
comment.go 定义评论相关模型。
*/
package model

import "time"

// Comment 表示博客详情页展示的一级评论。
type Comment struct {
	ID          int64     `json:"id"`
	PostID      int64     `json:"postId"`
	UserID      int64     `json:"userId"`
	Username    string    `json:"username"`
	DisplayName string    `json:"displayName"`
	Content     string    `json:"content"`
	CreatedAt   time.Time `json:"createdAt"`
}

