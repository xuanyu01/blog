/*
这个文件实现博客相关的业务服务
*/
package service

import (
	"blog/model"
	"blog/repository"
	"database/sql"
	"errors"
	"strings"
)

// BlogService 负责组织博客相关业务能力
type BlogService struct {
	blogRepo *repository.BlogRepository
}

// NewBlogService 创建博客业务服务
func NewBlogService(blogRepo *repository.BlogRepository) *BlogService {
	return &BlogService{blogRepo: blogRepo}
}

// ListBlogs 返回博客列表
func (s *BlogService) ListBlogs() ([]model.Blog, error) {
	return s.blogRepo.List()
}

// CreateBlog 用户创建新博客
func (s *BlogService) CreateBlog(title, content, authorUsername string) error {
	title = strings.TrimSpace(title)
	content = strings.TrimSpace(content)
	authorUsername = strings.TrimSpace(authorUsername)

	if authorUsername == "" {
		return errors.New("unauthorized")
	}
	if title == "" || content == "" {
		return errors.New("title and content are required")
	}

	// 这里把长度校验放在服务层
	// 这样无论从哪个入口创建博客 都会遵守同一套规则
	if len(title) > 50 {
		return errors.New("title cannot be longer than 50 characters")
	}
	if len(content) > 5000 {
		return errors.New("content cannot be longer than 5000 characters")
	}

	blog := model.Blog{
		Title:          title,
		Content:        content,
		AuthorUsername: authorUsername,
	}

	return s.blogRepo.Create(&blog)
}

// DeleteBlog 删除指定博客
func (s *BlogService) DeleteBlog(blogID int64, currentUsername, currentPermission string) error {
	currentUsername = strings.TrimSpace(currentUsername)
	if currentUsername == "" {
		return errors.New("unauthorized")
	}

	authorUsername, err := s.blogRepo.GetAuthorByID(blogID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("blog not found")
		}
		return err
	}

	if authorUsername != currentUsername && !model.CanManageAllBlogs(strings.TrimSpace(currentPermission)) {
		return errors.New("only the author can delete this blog")
	}

	return s.blogRepo.Delete(blogID)
}
