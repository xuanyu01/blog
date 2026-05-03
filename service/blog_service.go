/*
blog_service.go 负责博客、分类、标签、归档和互动相关业务逻辑。
*/
package service

import (
	"blog/model"
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
)

// BlogService 负责编排博客相关业务能力。
type BlogService struct {
	blogRepo blogRepository
}

type blogRepository interface {
	List(page, pageSize int, query model.BlogListQuery) (*model.BlogListResult, error)
	AdminList(page, pageSize int, keyword, author, status string) (*model.BlogListResult, error)
	ListByAuthor(page, pageSize int, authorUsername, status string) (*model.BlogListResult, error)
	ListFavoritesByUser(page, pageSize int, username string) (*model.BlogListResult, error)
	ListLikesByUser(page, pageSize int, username string) (*model.BlogListResult, error)
	GetByID(blogID int64) (*model.Blog, error)
	Create(blog *model.Blog) error
	GetAuthorByID(blogID int64) (string, error)
	Update(blog *model.Blog) error
	Review(blogID int64, status string, isTop bool) error
	Delete(blogID int64) error
	ListCategories() ([]model.Category, error)
	ListCategoriesForManage() ([]model.Category, error)
	CreateCategory(category *model.Category) error
	UpdateCategory(category *model.Category) error
	HideCategory(categoryID int64) error
	ListTags() ([]model.Tag, error)
	ListArchives() ([]model.ArchiveItem, error)
	IncrementViewCount(blogID int64) error
	HasLiked(blogID int64, username string) (bool, error)
	HasFavorited(blogID int64, username string) (bool, error)
	ToggleLike(blogID int64, username string) (bool, int64, error)
	ToggleFavorite(blogID int64, username string) (bool, int64, error)
}

// NewBlogService 创建博客服务。
func NewBlogService(blogRepo blogRepository) *BlogService {
	return &BlogService{blogRepo: blogRepo}
}

// ListBlogs 返回前台博客列表。
func (s *BlogService) ListBlogs(page, pageSize int, query model.BlogListQuery) (*model.BlogListResult, error) {
	page, pageSize = normalizeBlogPagination(page, pageSize)
	query.Keyword = strings.TrimSpace(query.Keyword)
	query.Tag = strings.TrimSpace(query.Tag)
	query.Archive = strings.TrimSpace(query.Archive)
	if query.Archive != "" && !isValidArchive(query.Archive) {
		return nil, errors.New("invalid archive")
	}
	return s.blogRepo.List(page, pageSize, query)
}

// ListManagedBlogs 返回后台博客列表。
func (s *BlogService) ListManagedBlogs(page, pageSize int, keyword, author, status, currentPermission string) (*model.BlogListResult, error) {
	if !model.CanManageAllBlogs(strings.TrimSpace(currentPermission)) {
		return nil, errors.New("forbidden")
	}

	page, pageSize = normalizeBlogPagination(page, pageSize)
	status = strings.TrimSpace(status)
	if status != "" && status != "draft" && status != "published" && status != "hidden" {
		return nil, errors.New("invalid status")
	}

	return s.blogRepo.AdminList(page, pageSize, keyword, author, status)
}

// ListCurrentUserBlogs 返回当前用户自己的博客列表。
func (s *BlogService) ListCurrentUserBlogs(page, pageSize int, status, currentUsername string) (*model.BlogListResult, error) {
	currentUsername = strings.TrimSpace(currentUsername)
	if currentUsername == "" {
		return nil, errors.New("unauthorized")
	}

	page, pageSize = normalizeBlogPagination(page, pageSize)
	status = strings.TrimSpace(status)
	if status != "" && status != "draft" && status != "published" && status != "hidden" {
		return nil, errors.New("invalid status")
	}

	return s.blogRepo.ListByAuthor(page, pageSize, currentUsername, status)
}

// ListFavoriteBlogs 返回当前用户收藏的博客列表。
func (s *BlogService) ListFavoriteBlogs(page, pageSize int, currentUsername string) (*model.BlogListResult, error) {
	currentUsername = strings.TrimSpace(currentUsername)
	if currentUsername == "" {
		return nil, errors.New("unauthorized")
	}

	page, pageSize = normalizeBlogPagination(page, pageSize)
	return s.blogRepo.ListFavoritesByUser(page, pageSize, currentUsername)
}

// ListLikedBlogs 返回当前用户点赞过的博客列表。
func (s *BlogService) ListLikedBlogs(page, pageSize int, currentUsername string) (*model.BlogListResult, error) {
	currentUsername = strings.TrimSpace(currentUsername)
	if currentUsername == "" {
		return nil, errors.New("unauthorized")
	}

	page, pageSize = normalizeBlogPagination(page, pageSize)
	return s.blogRepo.ListLikesByUser(page, pageSize, currentUsername)
}

// GetBlogByID 读取博客详情。
func (s *BlogService) GetBlogByID(blogID int64) (*model.Blog, error) {
	return s.GetBlogByIDForUser(blogID, "", "")
}

// GetBlogByIDForUser 按当前用户权限读取博客详情并增加阅读量。
func (s *BlogService) GetBlogByIDForUser(blogID int64, currentUsername, currentPermission string) (*model.Blog, error) {
	blog, err := s.blogRepo.GetByID(blogID)
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

	if err := s.blogRepo.IncrementViewCount(blogID); err == nil {
		blog.Stats.ViewCount++
	}

	if currentUsername != "" {
		if liked, err := s.blogRepo.HasLiked(blogID, currentUsername); err == nil {
			blog.Liked = liked
		}
		if favorited, err := s.blogRepo.HasFavorited(blogID, currentUsername); err == nil {
			blog.Favorited = favorited
		}
	}

	return blog, nil
}

// CreateBlog 创建博客并返回新文章。
func (s *BlogService) CreateBlog(input model.BlogCreateInput) (*model.Blog, error) {
	input.AuthorUsername = strings.TrimSpace(input.AuthorUsername)
	if input.AuthorUsername == "" {
		return nil, errors.New("unauthorized")
	}

	status, err := normalizeEditableStatus(input.Status)
	if err != nil {
		return nil, err
	}

	title, content, err := validateBlogInput(input.Title, input.Content)
	if err != nil {
		return nil, err
	}

	tags, err := normalizeTagNames(input.Tags)
	if err != nil {
		return nil, err
	}

	blog := model.Blog{
		Slug:           buildBlogSlug(title),
		Summary:        buildBlogSummary(content),
		Title:          title,
		Content:        content,
		AuthorUsername: input.AuthorUsername,
		Status:         status,
		IsTop:          model.CanManageAllBlogs(strings.TrimSpace(input.Permission)) && input.IsTop,
		CategoryID:     input.CategoryID,
		Tags:           tags,
	}

	if err := s.blogRepo.Create(&blog); err != nil {
		if err.Error() == "category not found" {
			return nil, err
		}
		return nil, err
	}

	return &blog, nil
}

// UpdateBlog 更新博客内容和状态。
func (s *BlogService) UpdateBlog(input model.BlogUpdateInput) error {
	input.CurrentUser = strings.TrimSpace(input.CurrentUser)
	input.CurrentPerm = strings.TrimSpace(input.CurrentPerm)
	if input.CurrentUser == "" {
		return errors.New("unauthorized")
	}

	status, err := normalizeEditableStatus(input.Status)
	if err != nil {
		return err
	}

	title, content, err := validateBlogInput(input.Title, input.Content)
	if err != nil {
		return err
	}

	tags, err := normalizeTagNames(input.Tags)
	if err != nil {
		return err
	}

	blog, err := s.blogRepo.GetByID(input.BlogID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("blog not found")
		}
		return err
	}

	if blog.AuthorUsername != input.CurrentUser && !model.CanManageAllBlogs(input.CurrentPerm) {
		return errors.New("only the author can edit this blog")
	}

	blog.Title = title
	blog.Content = content
	blog.Summary = buildBlogSummary(content)
	blog.Status = status
	blog.CategoryID = input.CategoryID
	blog.Tags = tags
	if status == "published" && blog.PublishedAt == nil {
		now := time.Now()
		blog.PublishedAt = &now
	}
	if status != "published" {
		blog.PublishedAt = nil
	}
	if model.CanManageAllBlogs(input.CurrentPerm) {
		blog.IsTop = input.IsTop
	}

	return s.blogRepo.Update(blog)
}

// ReviewBlog 审核博客状态。
func (s *BlogService) ReviewBlog(blogID int64, status string, isTop bool, currentPermission string) error {
	if !model.CanManageAllBlogs(strings.TrimSpace(currentPermission)) {
		return errors.New("forbidden")
	}

	status = strings.TrimSpace(status)
	if status != "draft" && status != "published" && status != "hidden" {
		return errors.New("invalid status")
	}

	if err := s.blogRepo.Review(blogID, status, isTop); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("blog not found")
		}
		return err
	}

	return nil
}

// DeleteBlog 删除博客。
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

// ListCategories 返回分类列表。
func (s *BlogService) ListCategories() ([]model.Category, error) {
	return s.blogRepo.ListCategories()
}

// ListManageCategories 返回后台分类列表。
func (s *BlogService) ListManageCategories(currentPermission string) ([]model.Category, error) {
	if !model.CanManageAllBlogs(strings.TrimSpace(currentPermission)) {
		return nil, errors.New("forbidden")
	}
	return s.blogRepo.ListCategoriesForManage()
}

// CreateCategory 创建新分类。
func (s *BlogService) CreateCategory(name, currentPermission string) (*model.Category, error) {
	if !model.CanManageAllBlogs(strings.TrimSpace(currentPermission)) {
		return nil, errors.New("forbidden")
	}

	name = strings.TrimSpace(name)
	if name == "" {
		return nil, errors.New("category name is required")
	}
	if len([]rune(name)) > 50 {
		return nil, errors.New("category name cannot be longer than 50 characters")
	}

	category := &model.Category{
		Name:   name,
		Slug:   buildTagSlug(name),
		Status: "active",
	}
	if category.Slug == "" {
		return nil, errors.New("invalid category")
	}

	if err := s.blogRepo.CreateCategory(category); err != nil {
		return nil, err
	}
	return category, nil
}

// UpdateCategory 更新分类名称。
func (s *BlogService) UpdateCategory(categoryID int64, name, currentPermission string) (*model.Category, error) {
	if !model.CanManageAllBlogs(strings.TrimSpace(currentPermission)) {
		return nil, errors.New("forbidden")
	}
	if categoryID <= 0 {
		return nil, errors.New("invalid category id")
	}

	name = strings.TrimSpace(name)
	if name == "" {
		return nil, errors.New("category name is required")
	}
	if len([]rune(name)) > 50 {
		return nil, errors.New("category name cannot be longer than 50 characters")
	}

	category := &model.Category{
		ID:     categoryID,
		Name:   name,
		Slug:   buildTagSlug(name),
		Status: "active",
	}
	if category.Slug == "" {
		return nil, errors.New("invalid category")
	}

	if err := s.blogRepo.UpdateCategory(category); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("category not found")
		}
		return nil, err
	}
	return category, nil
}

// DeleteCategory 隐藏分类，使其不再出现在可选列表中。
func (s *BlogService) DeleteCategory(categoryID int64, currentPermission string) error {
	if !model.CanManageAllBlogs(strings.TrimSpace(currentPermission)) {
		return errors.New("forbidden")
	}
	if categoryID <= 0 {
		return errors.New("invalid category id")
	}

	if err := s.blogRepo.HideCategory(categoryID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("category not found")
		}
		return err
	}
	return nil
}

// ListTags 返回标签列表。
func (s *BlogService) ListTags() ([]model.Tag, error) {
	return s.blogRepo.ListTags()
}

// ListArchives 返回归档列表。
func (s *BlogService) ListArchives() ([]model.ArchiveItem, error) {
	return s.blogRepo.ListArchives()
}

// ToggleLike 切换点赞状态。
func (s *BlogService) ToggleLike(blogID int64, currentUsername, currentPermission string) (*model.BlogInteraction, error) {
	if _, err := s.getInteractiveBlog(blogID, currentUsername, currentPermission); err != nil {
		return nil, err
	}
	if strings.TrimSpace(currentUsername) == "" {
		return nil, errors.New("unauthorized")
	}

	active, count, err := s.blogRepo.ToggleLike(blogID, currentUsername)
	if err != nil {
		return nil, err
	}

	blog, err := s.blogRepo.GetByID(blogID)
	if err != nil {
		return nil, err
	}

	return &model.BlogInteraction{
		Active:        active,
		LikeCount:     count,
		FavoriteCount: blog.Stats.FavoriteCount,
	}, nil
}

// ToggleFavorite 切换收藏状态。
func (s *BlogService) ToggleFavorite(blogID int64, currentUsername, currentPermission string) (*model.BlogInteraction, error) {
	if _, err := s.getInteractiveBlog(blogID, currentUsername, currentPermission); err != nil {
		return nil, err
	}
	if strings.TrimSpace(currentUsername) == "" {
		return nil, errors.New("unauthorized")
	}

	active, count, err := s.blogRepo.ToggleFavorite(blogID, currentUsername)
	if err != nil {
		return nil, err
	}

	blog, err := s.blogRepo.GetByID(blogID)
	if err != nil {
		return nil, err
	}

	return &model.BlogInteraction{
		Active:        active,
		LikeCount:     blog.Stats.LikeCount,
		FavoriteCount: count,
	}, nil
}

// getInteractiveBlog 读取博客并校验当前用户是否可访问。
func (s *BlogService) getInteractiveBlog(blogID int64, currentUsername, currentPermission string) (*model.Blog, error) {
	blog, err := s.blogRepo.GetByID(blogID)
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

// normalizeBlogPagination 规范分页参数范围。
func normalizeBlogPagination(page, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 50 {
		pageSize = 50
	}
	return page, pageSize
}

// normalizeEditableStatus 规范可编辑博客的状态字段。
func normalizeEditableStatus(status string) (string, error) {
	status = strings.TrimSpace(status)
	if status == "" {
		return "draft", nil
	}
	if status != "draft" && status != "published" {
		return "", errors.New("invalid status")
	}
	return status, nil
}

// validateBlogInput 清洗并校验标题与正文内容。
func validateBlogInput(title, content string) (string, string, error) {
	title = strings.TrimSpace(title)
	content = strings.TrimSpace(content)

	if title == "" || content == "" {
		return "", "", errors.New("title and content are required")
	}
	if len([]rune(title)) > 100 {
		return "", "", errors.New("title cannot be longer than 100 characters")
	}
	if len([]rune(content)) > 20000 {
		return "", "", errors.New("content cannot be longer than 20000 characters")
	}

	return title, content, nil
}

// normalizeTagNames 清洗标签名并生成对应标签结构。
func normalizeTagNames(raw []string) ([]model.Tag, error) {
	if len(raw) == 0 {
		return nil, nil
	}

	seen := map[string]struct{}{}
	items := make([]model.Tag, 0, len(raw))
	for _, name := range raw {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		if len([]rune(name)) > 20 {
			return nil, errors.New("tag cannot be longer than 20 characters")
		}

		slug := buildTagSlug(name)
		if slug == "" {
			return nil, errors.New("invalid tag")
		}
		if _, ok := seen[slug]; ok {
			continue
		}
		seen[slug] = struct{}{}

		items = append(items, model.Tag{
			Name: name,
			Slug: slug,
		})
	}

	if len(items) > 5 {
		return nil, errors.New("cannot set more than 5 tags")
	}

	return items, nil
}

// buildBlogSlug 根据标题生成 URL 友好的别名。
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

// buildBlogSummary 从正文中提取摘要内容。
func buildBlogSummary(content string) string {
	content = buildBlogPlainText(content)
	runes := []rune(content)
	if len(runes) <= 180 {
		return content
	}
	return string(runes[:180])
}

// buildBlogPlainText 把 Markdown 正文压缩为纯文本。
func buildBlogPlainText(content string) string {
	replacer := strings.NewReplacer(
		"\r", " ",
		"\n", " ",
		"\t", " ",
		"#", " ",
		"*", " ",
		"_", " ",
		"`", " ",
		">", " ",
		"-", " ",
		"|", " ",
		"[", " ",
		"]", " ",
		"(", " ",
		")", " ",
		"!", " ",
	)
	content = replacer.Replace(content)
	return strings.Join(strings.Fields(content), " ")
}

// buildTagSlug 根据标签名生成标签别名。
func buildTagSlug(name string) string {
	name = strings.TrimSpace(strings.ToLower(name))
	re := regexp.MustCompile(`[^a-z0-9]+`)
	slug := re.ReplaceAllString(name, "-")
	slug = strings.Trim(slug, "-")
	if slug != "" {
		return slug
	}

	var builder strings.Builder
	for _, r := range name {
		if r > 127 {
			builder.WriteRune(r)
		}
	}
	return strings.TrimSpace(builder.String())
}

// isValidArchive 校验归档参数是否符合 YYYY-MM 格式。
func isValidArchive(value string) bool {
	re := regexp.MustCompile(`^\d{4}-\d{2}$`)
	if !re.MatchString(value) {
		return false
	}

	month := value[5:]
	return month >= "01" && month <= "12"
}
