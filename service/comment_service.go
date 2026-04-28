/*
comment_service.go 负责评论相关业务逻辑。
*/
package service

import (
	"blog/model"
	"database/sql"
	"errors"
	"strings"
)

// CommentService 负责评论业务编排。
type CommentService struct {
	commentRepo commentRepository
	blogRepo    commentBlogRepository
}

type commentRepository interface {
	ListByPostID(postID int64) ([]model.Comment, error)
	Create(postID int64, username, content string) (*model.Comment, error)
	GetAuthorByID(commentID int64) (string, error)
	Delete(commentID int64) error
}

type commentBlogRepository interface {
	GetByID(blogID int64) (*model.Blog, error)
}

// NewCommentService 创建评论服务。
func NewCommentService(commentRepo commentRepository, blogRepo commentBlogRepository) *CommentService {
	return &CommentService{
		commentRepo: commentRepo,
		blogRepo:    blogRepo,
	}
}

// ListComments 返回当前用户可见文章下的一级评论。
func (s *CommentService) ListComments(postID int64, currentUsername, currentPermission string) ([]model.Comment, error) {
	if _, err := s.getAccessibleBlog(postID, currentUsername, currentPermission); err != nil {
		return nil, err
	}
	return s.commentRepo.ListByPostID(postID)
}

// CreateComment 创建一级评论。
func (s *CommentService) CreateComment(postID int64, content, currentUsername, currentPermission string) (*model.Comment, error) {
	currentUsername = strings.TrimSpace(currentUsername)
	if currentUsername == "" {
		return nil, errors.New("unauthorized")
	}

	if _, err := s.getAccessibleBlog(postID, currentUsername, currentPermission); err != nil {
		return nil, err
	}

	content = strings.TrimSpace(content)
	if content == "" {
		return nil, errors.New("content is required")
	}
	if len([]rune(content)) > 500 {
		return nil, errors.New("content cannot be longer than 500 characters")
	}

	return s.commentRepo.Create(postID, currentUsername, content)
}

// DeleteComment 删除自己评论，管理员可删除任意评论。
func (s *CommentService) DeleteComment(commentID int64, currentUsername, currentPermission string) error {
	currentUsername = strings.TrimSpace(currentUsername)
	currentPermission = strings.TrimSpace(currentPermission)
	if currentUsername == "" {
		return errors.New("unauthorized")
	}

	authorUsername, err := s.commentRepo.GetAuthorByID(commentID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("comment not found")
		}
		return err
	}

	if authorUsername != currentUsername && !model.CanManageAllBlogs(currentPermission) {
		return errors.New("forbidden")
	}

	if err := s.commentRepo.Delete(commentID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("comment not found")
		}
		return err
	}

	return nil
}

// getAccessibleBlog 校验当前用户是否有权访问目标博客。
func (s *CommentService) getAccessibleBlog(postID int64, currentUsername, currentPermission string) (*model.Blog, error) {
	blog, err := s.blogRepo.GetByID(postID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("blog not found")
		}
		return nil, err
	}

	currentUsername = strings.TrimSpace(currentUsername)
	currentPermission = strings.TrimSpace(currentPermission)
	if blog.Status != "published" && blog.AuthorUsername != currentUsername && !model.CanManageAllBlogs(currentPermission) {
		return nil, errors.New("forbidden")
	}

	return blog, nil
}
