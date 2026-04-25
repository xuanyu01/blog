/*
这个文件实现博客相关的业务服务
*/
package service

import (
	"blog/model"
	"blog/repository"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
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
func (s *BlogService) ListBlogs(page, pageSize int, keyword string) (*model.BlogListResult, error) {
	if page < 1 {
		page = 1
	}

	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 50 {
		pageSize = 50
	}

	return s.blogRepo.List(page, pageSize, keyword)
}

// GetBlogByID 返回指定博客详情
func (s *BlogService) GetBlogByID(blogID int64) (*model.Blog, error) {
	blog, err := s.blogRepo.GetByID(blogID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("blog not found")
		}
		return nil, err
	}

	return blog, nil
}

// CreateBlog 用户创建新博客
func (s *BlogService) CreateBlog(title, content, authorUsername string) error {
	authorUsername = strings.TrimSpace(authorUsername)

	if authorUsername == "" {
		return errors.New("unauthorized")
	}

	title, content, err := validateBlogInput(title, content)
	if err != nil {
		return err
	}

	blog := model.Blog{
		Slug:           buildBlogSlug(title),
		Summary:        buildBlogSummary(content),
		Title:          title,
		Content:        content,
		AuthorUsername: authorUsername,
	}

	return s.blogRepo.Create(&blog)
}

// UpdateBlog 更新指定博客内容 仅作者或管理员可操作
func (s *BlogService) UpdateBlog(blogID int64, title, content, currentUsername, currentPermission string) error {
	currentUsername = strings.TrimSpace(currentUsername)
	currentPermission = strings.TrimSpace(currentPermission)
	if currentUsername == "" {
		return errors.New("unauthorized")
	}

	title, content, err := validateBlogInput(title, content)
	if err != nil {
		return err
	}

	blog, err := s.blogRepo.GetByID(blogID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("blog not found")
		}
		return err
	}

	if blog.AuthorUsername != currentUsername && !model.CanManageAllBlogs(currentPermission) {
		return errors.New("only the author can edit this blog")
	}

	blog.Title = title
	blog.Content = content
	blog.Summary = buildBlogSummary(content)
	return s.blogRepo.Update(blog)
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

func validateBlogInput(title, content string) (string, string, error) {
	title = strings.TrimSpace(title)
	content = strings.TrimSpace(content)

	if title == "" || content == "" {
		return "", "", errors.New("title and content are required")
	}

	// 这里把长度校验放在服务层
	// 这样无论从哪个入口创建或编辑博客 都会遵守同一套规则
	if len(title) > 50 {
		return "", "", errors.New("title cannot be longer than 50 characters")
	}
	if len(content) > 5000 {
		return "", "", errors.New("content cannot be longer than 5000 characters")
	}

	return title, content, nil
}

func buildBlogSlug(title string) string {
	title = strings.TrimSpace(strings.ToLower(title))
	var builder strings.Builder
	lastDash := false

	for _, r := range title {
		switch {
		case r >= 'a' && r <= 'z':
			builder.WriteRune(r)
			lastDash = false
		case r >= '0' && r <= '9':
			builder.WriteRune(r)
			lastDash = false
		default:
			if !lastDash && builder.Len() > 0 {
				builder.WriteByte('-')
				lastDash = true
			}
		}
	}

	slug := strings.Trim(builder.String(), "-")
	if slug == "" {
		slug = "post"
	}

	return fmt.Sprintf("%s-%d", slug, time.Now().UnixNano())
}

func buildBlogSummary(content string) string {
	content = strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(content, "\r", " "), "\n", " "))
	runes := []rune(content)
	if len(runes) <= 180 {
		return content
	}
	return string(runes[:180])
}
