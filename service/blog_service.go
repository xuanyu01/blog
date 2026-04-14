/*
这个文件实现博客相关的业务服务
*/
package service

import (
	"blog/model"
	"blog/repository"
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
