package service

import (
	"blog/model"
	"database/sql"
	"testing"
)

type fakeBlogRepo struct {
	listPage        int
	listPageSize    int
	listQuery       model.BlogListQuery
	adminListPage   int
	adminListSize   int
	adminListKW     string
	adminListAuthor string
	adminListStatus string
	favoritePage    int
	favoriteSize    int
	favoriteUser    string
	listResult      *model.BlogListResult
	adminListResult *model.BlogListResult
	favoriteResult  *model.BlogListResult
	blog            *model.Blog
	createBlog      *model.Blog
	updateBlog      *model.Blog
	author          string
	deleteID        int64
	reviewID        int64
	reviewStatus    string
	reviewIsTop     bool
	categories      []model.Category
	tags            []model.Tag
	archives        []model.ArchiveItem
	incrementViewID int64
	hasLiked        bool
	hasFavorited    bool
	toggleLikeOn    bool
	toggleLikeCount int64
	toggleFavOn     bool
	toggleFavCount  int64
}

func (f *fakeBlogRepo) List(page, pageSize int, query model.BlogListQuery) (*model.BlogListResult, error) {
	f.listPage = page
	f.listPageSize = pageSize
	f.listQuery = query
	if f.listResult != nil {
		return f.listResult, nil
	}
	return &model.BlogListResult{}, nil
}

func (f *fakeBlogRepo) AdminList(page, pageSize int, keyword, author, status string) (*model.BlogListResult, error) {
	f.adminListPage = page
	f.adminListSize = pageSize
	f.adminListKW = keyword
	f.adminListAuthor = author
	f.adminListStatus = status
	if f.adminListResult != nil {
		return f.adminListResult, nil
	}
	return &model.BlogListResult{}, nil
}

func (f *fakeBlogRepo) ListByAuthor(page, pageSize int, authorUsername, status string) (*model.BlogListResult, error) {
	f.adminListPage = page
	f.adminListSize = pageSize
	f.adminListAuthor = authorUsername
	f.adminListStatus = status
	if f.adminListResult != nil {
		return f.adminListResult, nil
	}
	return &model.BlogListResult{}, nil
}

func (f *fakeBlogRepo) ListFavoritesByUser(page, pageSize int, username string) (*model.BlogListResult, error) {
	f.favoritePage = page
	f.favoriteSize = pageSize
	f.favoriteUser = username
	if f.favoriteResult != nil {
		return f.favoriteResult, nil
	}
	return &model.BlogListResult{}, nil
}

func (f *fakeBlogRepo) GetByID(blogID int64) (*model.Blog, error) {
	if f.blog == nil {
		return nil, sql.ErrNoRows
	}
	copy := *f.blog
	return &copy, nil
}

func (f *fakeBlogRepo) Create(blog *model.Blog) error {
	copy := *blog
	f.createBlog = &copy
	return nil
}

func (f *fakeBlogRepo) GetAuthorByID(blogID int64) (string, error) {
	if f.author == "" {
		return "", sql.ErrNoRows
	}
	return f.author, nil
}

func (f *fakeBlogRepo) Update(blog *model.Blog) error {
	copy := *blog
	f.updateBlog = &copy
	return nil
}

func (f *fakeBlogRepo) Review(blogID int64, status string, isTop bool) error {
	f.reviewID = blogID
	f.reviewStatus = status
	f.reviewIsTop = isTop
	if blogID == 404 {
		return sql.ErrNoRows
	}
	return nil
}

func (f *fakeBlogRepo) Delete(blogID int64) error {
	f.deleteID = blogID
	return nil
}

func (f *fakeBlogRepo) ListCategories() ([]model.Category, error) {
	return f.categories, nil
}

func (f *fakeBlogRepo) ListCategoriesForManage() ([]model.Category, error) {
	return f.categories, nil
}

func (f *fakeBlogRepo) CreateCategory(category *model.Category) error {
	category.ID = 99
	return nil
}

func (f *fakeBlogRepo) UpdateCategory(category *model.Category) error {
	if category.ID == 404 {
		return sql.ErrNoRows
	}
	return nil
}

func (f *fakeBlogRepo) HideCategory(categoryID int64) error {
	if categoryID == 404 {
		return sql.ErrNoRows
	}
	return nil
}

func (f *fakeBlogRepo) ListTags() ([]model.Tag, error) {
	return f.tags, nil
}

func (f *fakeBlogRepo) ListArchives() ([]model.ArchiveItem, error) {
	return f.archives, nil
}

func (f *fakeBlogRepo) IncrementViewCount(blogID int64) error {
	f.incrementViewID = blogID
	return nil
}

func (f *fakeBlogRepo) HasLiked(blogID int64, username string) (bool, error) {
	return f.hasLiked, nil
}

func (f *fakeBlogRepo) HasFavorited(blogID int64, username string) (bool, error) {
	return f.hasFavorited, nil
}

func (f *fakeBlogRepo) ToggleLike(blogID int64, username string) (bool, int64, error) {
	return f.toggleLikeOn, f.toggleLikeCount, nil
}

func (f *fakeBlogRepo) ToggleFavorite(blogID int64, username string) (bool, int64, error) {
	return f.toggleFavOn, f.toggleFavCount, nil
}

func TestBlogServiceCreateBlogBuildsSlugSummaryStatusAndTags(t *testing.T) {
	repo := &fakeBlogRepo{}
	service := NewBlogService(repo)

	blog, err := service.CreateBlog(model.BlogCreateInput{
		Title:          "Hello World",
		Content:        "first line\nsecond line",
		Status:         "draft",
		CategoryID:     2,
		Tags:           []string{"Go", "Gin", "Go"},
		AuthorUsername: "alice",
		Permission:     model.PermissionUser,
	})
	if err != nil {
		t.Fatalf("CreateBlog returned error: %v", err)
	}

	if repo.createBlog == nil {
		t.Fatal("expected repository Create to be called")
	}
	if repo.createBlog.Slug == "" || repo.createBlog.Summary == "" {
		t.Fatalf("expected slug and summary to be generated, got slug=%q summary=%q", repo.createBlog.Slug, repo.createBlog.Summary)
	}
	if repo.createBlog.Status != "draft" || repo.createBlog.CategoryID != 2 {
		t.Fatalf("unexpected create payload: status=%q category=%d", repo.createBlog.Status, repo.createBlog.CategoryID)
	}
	if len(repo.createBlog.Tags) != 2 {
		t.Fatalf("expected 2 normalized tags, got %d", len(repo.createBlog.Tags))
	}
	if blog == nil {
		t.Fatal("expected created blog to be returned")
	}
}

func TestBlogServiceUpdateBlogRequiresAuthorOrAdmin(t *testing.T) {
	repo := &fakeBlogRepo{
		blog: &model.Blog{ID: 1, Title: "Old", Content: "Old", AuthorUsername: "alice", Status: "draft"},
	}
	service := NewBlogService(repo)

	err := service.UpdateBlog(model.BlogUpdateInput{
		BlogID:      1,
		Title:       "New",
		Content:     "New Content",
		Status:      "published",
		CurrentUser: "bob",
		CurrentPerm: model.PermissionUser,
	})
	if err == nil || err.Error() != "only the author can edit this blog" {
		t.Fatalf("expected author permission error, got %v", err)
	}
}

func TestBlogServiceUpdateBlogAllowsPublishAndTopForManager(t *testing.T) {
	repo := &fakeBlogRepo{
		blog: &model.Blog{ID: 2, Title: "Old", Content: "Old", AuthorUsername: "alice", Status: "draft"},
	}
	service := NewBlogService(repo)

	err := service.UpdateBlog(model.BlogUpdateInput{
		BlogID:      2,
		Title:       "New",
		Content:     "New Content",
		Status:      "published",
		IsTop:       true,
		CurrentUser: "alice",
		CurrentPerm: model.PermissionAdmin,
		Tags:        []string{"go"},
	})
	if err != nil {
		t.Fatalf("UpdateBlog returned error: %v", err)
	}
	if repo.updateBlog == nil {
		t.Fatal("expected repository Update to be called")
	}
	if repo.updateBlog.Status != "published" || !repo.updateBlog.IsTop || repo.updateBlog.PublishedAt == nil {
		t.Fatalf("expected published top blog with publishedAt, got status=%q isTop=%v publishedAt=%v", repo.updateBlog.Status, repo.updateBlog.IsTop, repo.updateBlog.PublishedAt)
	}
	if len(repo.updateBlog.Tags) != 1 {
		t.Fatalf("expected tags to be normalized, got %d", len(repo.updateBlog.Tags))
	}
}

func TestBlogServiceGetBlogByIDForUserBlocksDraftForOthers(t *testing.T) {
	repo := &fakeBlogRepo{
		blog: &model.Blog{ID: 1, AuthorUsername: "alice", Status: "draft"},
	}
	service := NewBlogService(repo)

	_, err := service.GetBlogByIDForUser(1, "bob", model.PermissionUser)
	if err == nil || err.Error() != "forbidden" {
		t.Fatalf("expected forbidden error, got %v", err)
	}
}

func TestBlogServiceGetBlogByIDForUserLoadsInteractionState(t *testing.T) {
	repo := &fakeBlogRepo{
		blog:         &model.Blog{ID: 8, AuthorUsername: "alice", Status: "published", Stats: model.BlogStats{ViewCount: 4}},
		hasLiked:     true,
		hasFavorited: true,
	}
	service := NewBlogService(repo)

	blog, err := service.GetBlogByIDForUser(8, "bob", model.PermissionUser)
	if err != nil {
		t.Fatalf("GetBlogByIDForUser returned error: %v", err)
	}
	if repo.incrementViewID != 8 {
		t.Fatalf("expected view count increment for blog 8, got %d", repo.incrementViewID)
	}
	if !blog.Liked || !blog.Favorited || blog.Stats.ViewCount != 5 {
		t.Fatalf("expected interaction state to be filled, got liked=%v favorited=%v views=%d", blog.Liked, blog.Favorited, blog.Stats.ViewCount)
	}
}

func TestBlogServiceDeleteBlogRequiresAuthorOrAdmin(t *testing.T) {
	repo := &fakeBlogRepo{author: "alice"}
	service := NewBlogService(repo)

	err := service.DeleteBlog(1, "bob", model.PermissionUser)
	if err == nil || err.Error() != "only the author can delete this blog" {
		t.Fatalf("expected delete permission error, got %v", err)
	}
}

func TestBlogServiceListBlogsNormalizesPaginationAndFilters(t *testing.T) {
	repo := &fakeBlogRepo{
		listResult: &model.BlogListResult{Page: 1, PageSize: 50, Total: 120, Keyword: "go"},
	}
	service := NewBlogService(repo)

	_, err := service.ListBlogs(0, 100, model.BlogListQuery{
		Keyword:    "go",
		CategoryID: 3,
		Tag:        "gin",
		Archive:    "2026-04",
	})
	if err != nil {
		t.Fatalf("ListBlogs returned error: %v", err)
	}

	if repo.listPage != 1 || repo.listPageSize != 50 || repo.listQuery.Tag != "gin" || repo.listQuery.CategoryID != 3 {
		t.Fatalf("expected normalized pagination and filters, got page=%d size=%d tag=%s category=%d", repo.listPage, repo.listPageSize, repo.listQuery.Tag, repo.listQuery.CategoryID)
	}
}

func TestBlogServiceListManagedBlogsRequiresManagerPermission(t *testing.T) {
	repo := &fakeBlogRepo{}
	service := NewBlogService(repo)

	_, err := service.ListManagedBlogs(1, 10, "", "", "", model.PermissionUser)
	if err == nil || err.Error() != "forbidden" {
		t.Fatalf("expected forbidden error, got %v", err)
	}
}

func TestBlogServiceListCurrentUserBlogsRequiresLogin(t *testing.T) {
	repo := &fakeBlogRepo{}
	service := NewBlogService(repo)

	_, err := service.ListCurrentUserBlogs(1, 10, "draft", "")
	if err == nil || err.Error() != "unauthorized" {
		t.Fatalf("expected unauthorized error, got %v", err)
	}
}

func TestBlogServiceListFavoriteBlogsRequiresLogin(t *testing.T) {
	repo := &fakeBlogRepo{}
	service := NewBlogService(repo)

	_, err := service.ListFavoriteBlogs(1, 10, "")
	if err == nil || err.Error() != "unauthorized" {
		t.Fatalf("expected unauthorized error, got %v", err)
	}
}

func TestBlogServiceListFavoriteBlogsCallsRepository(t *testing.T) {
	repo := &fakeBlogRepo{
		favoriteResult: &model.BlogListResult{Page: 1, PageSize: 10, Total: 2},
	}
	service := NewBlogService(repo)

	result, err := service.ListFavoriteBlogs(0, 100, "alice")
	if err != nil {
		t.Fatalf("ListFavoriteBlogs returned error: %v", err)
	}
	if repo.favoritePage != 1 || repo.favoriteSize != 50 || repo.favoriteUser != "alice" {
		t.Fatalf("expected normalized favorite query, got page=%d size=%d user=%s", repo.favoritePage, repo.favoriteSize, repo.favoriteUser)
	}
	if result.Total != 2 {
		t.Fatalf("expected favorite result total 2, got %d", result.Total)
	}
}

func TestBlogServiceReviewBlogValidatesStatus(t *testing.T) {
	repo := &fakeBlogRepo{}
	service := NewBlogService(repo)

	err := service.ReviewBlog(1, "archived", false, model.PermissionAdmin)
	if err == nil || err.Error() != "invalid status" {
		t.Fatalf("expected invalid status error, got %v", err)
	}
}

func TestBlogServiceReviewBlogCallsRepository(t *testing.T) {
	repo := &fakeBlogRepo{}
	service := NewBlogService(repo)

	if err := service.ReviewBlog(7, "published", true, model.PermissionAdmin); err != nil {
		t.Fatalf("ReviewBlog returned error: %v", err)
	}

	if repo.reviewID != 7 || repo.reviewStatus != "published" || !repo.reviewIsTop {
		t.Fatalf("expected review call to be recorded, got id=%d status=%s isTop=%v", repo.reviewID, repo.reviewStatus, repo.reviewIsTop)
	}
}

func TestBlogServiceToggleLikeRequiresLogin(t *testing.T) {
	repo := &fakeBlogRepo{
		blog: &model.Blog{ID: 1, AuthorUsername: "alice", Status: "published"},
	}
	service := NewBlogService(repo)

	_, err := service.ToggleLike(1, "", "")
	if err == nil || err.Error() != "unauthorized" {
		t.Fatalf("expected unauthorized error, got %v", err)
	}
}

func TestBlogServiceToggleFavoriteReturnsCounts(t *testing.T) {
	repo := &fakeBlogRepo{
		blog:           &model.Blog{ID: 2, AuthorUsername: "alice", Status: "published", Stats: model.BlogStats{LikeCount: 3}},
		toggleFavOn:    true,
		toggleFavCount: 6,
	}
	service := NewBlogService(repo)

	result, err := service.ToggleFavorite(2, "bob", model.PermissionUser)
	if err != nil {
		t.Fatalf("ToggleFavorite returned error: %v", err)
	}
	if !result.Active || result.FavoriteCount != 6 || result.LikeCount != 3 {
		t.Fatalf("unexpected toggle result: %+v", result)
	}
}

func TestBlogServiceCategoryManagementRequiresManager(t *testing.T) {
	repo := &fakeBlogRepo{}
	service := NewBlogService(repo)

	_, err := service.CreateCategory("Go", model.PermissionUser)
	if err == nil || err.Error() != "forbidden" {
		t.Fatalf("expected forbidden error, got %v", err)
	}
}

func TestBlogServiceCreateCategoryBuildsSlug(t *testing.T) {
	repo := &fakeBlogRepo{}
	service := NewBlogService(repo)

	category, err := service.CreateCategory("Go Web", model.PermissionAdmin)
	if err != nil {
		t.Fatalf("CreateCategory returned error: %v", err)
	}
	if category.ID != 99 || category.Slug != "go-web" {
		t.Fatalf("unexpected category result: %+v", category)
	}
}
