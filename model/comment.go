/*
定义评论相关模型。
*/
package model

import "time"

// Comment 表示博客评论；Replies 用于在一级评论下展示回复列表。
type Comment struct {
	ID              int64     `json:"id"`
	PostID          int64     `json:"postId"`
	UserID          int64     `json:"userId"`
	ParentID        *int64    `json:"parentId,omitempty"`
	RootID          *int64    `json:"rootId,omitempty"`
	Username        string    `json:"username"`
	DisplayName     string    `json:"displayName"`
	ReplyToUsername string    `json:"replyToUsername,omitempty"`
	Content         string    `json:"content"`
	CreatedAt       time.Time `json:"createdAt"`
	Replies         []Comment `json:"replies,omitempty"`
}
