/*
实现前端依赖的 HTTP 处理逻辑。
*/
package handler

import (
	"blog/model"
	"blog/service"
	"blog/session"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

var allowedAvatarTypes = map[string]string{
	".png":  "image/png",
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".gif":  "image/gif",
}

// WebHandler 负责博客和认证相关的 HTTP 处理。
type WebHandler struct {
	blogService    *service.BlogService
	commentService *service.CommentService
	authService    *service.AuthService
}

// appStateResponse 表示首页状态接口的响应结构。
type appStateResponse struct {
	User model.UserView `json:"user"`
}

// blogListResponse 表示博客列表响应结构。
type blogListResponse struct {
	Items      []model.Blog `json:"items"`
	Page       int          `json:"page"`
	PageSize   int          `json:"pageSize"`
	Total      int          `json:"total"`
	Keyword    string       `json:"keyword"`
	CategoryID int64        `json:"categoryId"`
	Tag        string       `json:"tag"`
	Archive    string       `json:"archive"`
}

// userListResponse 表示用户列表响应结构。
type userListResponse struct {
	Items    []model.UserListItem `json:"items"`
	Page     int                  `json:"page"`
	PageSize int                  `json:"pageSize"`
	Total    int                  `json:"total"`
}

// blogReviewRequest 表示博客审核请求。
type blogReviewRequest struct {
	Status string `json:"status"`
	IsTop  bool   `json:"isTop"`
}

// NewWebHandler 创建处理器实例。
type taxonomyListResponse[T any] struct {
	Items []T `json:"items"`
}

type blogInteractionResponse struct {
	Active        bool  `json:"active"`
	LikeCount     int64 `json:"likeCount"`
	FavoriteCount int64 `json:"favoriteCount"`
}

type categoryPayload struct {
	Name string `json:"name"`
}

type commentCreateRequest struct {
	Content string `json:"content"`
}

func NewWebHandler(blogService *service.BlogService, commentService *service.CommentService, authService *service.AuthService) *WebHandler {
	return &WebHandler{
		blogService:    blogService,
		commentService: commentService,
		authService:    authService,
	}
}

// GetAppState 返回首页所需的聚合状态。
func (h *WebHandler) GetAppState(c *gin.Context) {
	user := h.getCurrentUser(c)
	c.JSON(http.StatusOK, appStateResponse{
		User: user,
	})
}

// ListBlogs 返回前台博客列表。
func (h *WebHandler) ListBlogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	keyword := strings.TrimSpace(c.Query("keyword"))
	categoryID, _ := strconv.ParseInt(c.DefaultQuery("categoryId", "0"), 10, 64)
	archive := strings.TrimSpace(c.Query("archive"))

	result, err := h.blogService.ListBlogs(page, pageSize, model.BlogListQuery{
		Keyword:    keyword,
		CategoryID: categoryID,
		Archive:    archive,
	})
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "invalid archive" {
			statusCode = http.StatusBadRequest
		}
		log.Println("list blogs failed:", err)
		c.JSON(statusCode, gin.H{
			"message": "failed to load blogs",
		})
		return
	}

	c.JSON(http.StatusOK, blogListResponse{
		Items:      result.Items,
		Page:       result.Page,
		PageSize:   result.PageSize,
		Total:      result.Total,
		Keyword:    result.Keyword,
		CategoryID: result.CategoryID,
		Archive:    result.Archive,
	})
}

func (h *WebHandler) ListCategories(c *gin.Context) {
	items, err := h.blogService.ListCategories()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to load categories",
		})
		return
	}

	c.JSON(http.StatusOK, taxonomyListResponse[model.Category]{
		Items: items,
	})
}

func (h *WebHandler) ListTags(c *gin.Context) {
	items, err := h.blogService.ListTags()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to load tags",
		})
		return
	}

	c.JSON(http.StatusOK, taxonomyListResponse[model.Tag]{
		Items: items,
	})
}

func (h *WebHandler) ListArchives(c *gin.Context) {
	items, err := h.blogService.ListArchives()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to load archives",
		})
		return
	}

	c.JSON(http.StatusOK, taxonomyListResponse[model.ArchiveItem]{
		Items: items,
	})
}

func (h *WebHandler) ListManageCategories(c *gin.Context) {
	user := h.getCurrentUser(c)
	if !user.IsLogin {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "unauthorized"})
		return
	}

	items, err := h.blogService.ListManageCategories(user.Permission)
	if err != nil {
		statusCode := http.StatusBadRequest
		if err.Error() == "forbidden" {
			statusCode = http.StatusForbidden
		}
		c.JSON(statusCode, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, taxonomyListResponse[model.Category]{Items: items})
}

func (h *WebHandler) CreateCategory(c *gin.Context) {
	user := h.getCurrentUser(c)
	if !user.IsLogin {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "unauthorized"})
		return
	}

	var payload categoryPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid request body"})
		return
	}

	category, err := h.blogService.CreateCategory(payload.Name, user.Permission)
	if err != nil {
		statusCode := http.StatusBadRequest
		if err.Error() == "forbidden" {
			statusCode = http.StatusForbidden
		}
		c.JSON(statusCode, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"item": category})
}

func (h *WebHandler) UpdateCategory(c *gin.Context) {
	user := h.getCurrentUser(c)
	if !user.IsLogin {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "unauthorized"})
		return
	}

	categoryID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || categoryID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid category id"})
		return
	}

	var payload categoryPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid request body"})
		return
	}

	category, err := h.blogService.UpdateCategory(categoryID, payload.Name, user.Permission)
	if err != nil {
		statusCode := http.StatusBadRequest
		switch err.Error() {
		case "forbidden":
			statusCode = http.StatusForbidden
		case "category not found":
			statusCode = http.StatusNotFound
		}
		c.JSON(statusCode, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"item": category})
}

func (h *WebHandler) DeleteCategory(c *gin.Context) {
	user := h.getCurrentUser(c)
	if !user.IsLogin {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "unauthorized"})
		return
	}

	categoryID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || categoryID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid category id"})
		return
	}

	if err := h.blogService.DeleteCategory(categoryID, user.Permission); err != nil {
		statusCode := http.StatusBadRequest
		switch err.Error() {
		case "forbidden":
			statusCode = http.StatusForbidden
		case "category not found":
			statusCode = http.StatusNotFound
		}
		c.JSON(statusCode, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Category deleted"})
}

// ListManagedBlogs 返回后台博客列表。
func (h *WebHandler) ListManagedBlogs(c *gin.Context) {
	user := h.getCurrentUser(c)
	if !user.IsLogin {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "unauthorized",
		})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	keyword := strings.TrimSpace(c.Query("keyword"))
	author := strings.TrimSpace(c.Query("author"))
	status := strings.TrimSpace(c.Query("status"))

	result, err := h.blogService.ListManagedBlogs(page, pageSize, keyword, author, status, user.Permission)
	if err != nil {
		statusCode := http.StatusBadRequest
		switch err.Error() {
		case "unauthorized":
			statusCode = http.StatusUnauthorized
		case "forbidden":
			statusCode = http.StatusForbidden
		}

		c.JSON(statusCode, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, blogListResponse{
		Items:    result.Items,
		Page:     result.Page,
		PageSize: result.PageSize,
		Total:    result.Total,
		Keyword:  result.Keyword,
	})
}

// ListCurrentUserBlogs 返回当前用户自己的博客列表。
func (h *WebHandler) ListCurrentUserBlogs(c *gin.Context) {
	user := h.getCurrentUser(c)
	if !user.IsLogin {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "unauthorized",
		})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	status := strings.TrimSpace(c.Query("status"))

	result, err := h.blogService.ListCurrentUserBlogs(page, pageSize, status, user.UserName)
	if err != nil {
		statusCode := http.StatusBadRequest
		if err.Error() == "unauthorized" {
			statusCode = http.StatusUnauthorized
		}

		c.JSON(statusCode, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, blogListResponse{
		Items:    result.Items,
		Page:     result.Page,
		PageSize: result.PageSize,
		Total:    result.Total,
		Keyword:  result.Keyword,
	})
}

// ListFavoriteBlogs 返回当前用户的收藏列表。
func (h *WebHandler) ListFavoriteBlogs(c *gin.Context) {
	user := h.getCurrentUser(c)
	if !user.IsLogin {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "unauthorized",
		})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	result, err := h.blogService.ListFavoriteBlogs(page, pageSize, user.UserName)
	if err != nil {
		statusCode := http.StatusBadRequest
		if err.Error() == "unauthorized" {
			statusCode = http.StatusUnauthorized
		}

		c.JSON(statusCode, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, blogListResponse{
		Items:    result.Items,
		Page:     result.Page,
		PageSize: result.PageSize,
		Total:    result.Total,
	})
}

// Register 处理注册请求。
func (h *WebHandler) Register(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")

	if username == "" || password == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "username and password are required",
		})
		return
	}

	if err := h.authService.Register(username, password); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Register success",
	})
}

// Login 处理登录请求。
func (h *WebHandler) Login(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")

	sessionID, err := h.authService.Login(username, password)
	if err != nil {
		status := http.StatusBadRequest
		if !errors.Is(err, service.ErrInvalidCredentials) {
			status = http.StatusInternalServerError
			log.Println("login failed:", err)
		}

		c.JSON(status, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.SetCookie(session.CookieName, sessionID, int(session.Expire.Seconds()), "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{
		"message": "Login success",
	})
}

// Logout 处理退出登录请求。
func (h *WebHandler) Logout(c *gin.Context) {
	sessionID, _ := c.Cookie(session.CookieName)
	if err := h.authService.Logout(sessionID); err != nil {
		log.Println("logout failed:", err)
	}

	c.SetCookie(session.CookieName, "", -1, "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{
		"message": "Logout success",
	})
}

// CurrentUser 返回当前登录用户信息。
func (h *WebHandler) CurrentUser(c *gin.Context) {
	user := h.getCurrentUser(c)
	if !user.IsLogin {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "unauthorized",
		})
		return
	}

	c.JSON(http.StatusOK, user)
}

// UpdateProfile 处理用户资料修改请求。
func (h *WebHandler) UpdateProfile(c *gin.Context) {
	sessionID, err := c.Cookie(session.CookieName)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "unauthorized",
		})
		return
	}

	var payload model.UserProfileUpdate
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid request body",
		})
		return
	}

	user, err := h.authService.UpdateProfile(sessionID, payload)
	if err != nil {
		status := http.StatusBadRequest
		if err.Error() == "unauthorized" {
			status = http.StatusUnauthorized
		}
		c.JSON(status, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, user)
}

// UploadAvatar 处理头像上传请求。
func (h *WebHandler) UploadAvatar(c *gin.Context) {
	sessionID, err := c.Cookie(session.CookieName)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "unauthorized",
		})
		return
	}

	fileHeader, err := c.FormFile("avatar")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "avatar file is required",
		})
		return
	}

	fileName, err := saveAvatarFile(fileHeader)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	user, err := h.authService.UpdateAvatar(sessionID, fileName)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, user)
}

// UpdatePassword 处理密码修改请求。
func (h *WebHandler) UpdatePassword(c *gin.Context) {
	sessionID, err := c.Cookie(session.CookieName)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "unauthorized",
		})
		return
	}

	var payload model.PasswordUpdate
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid request body",
		})
		return
	}

	if err := h.authService.UpdatePassword(sessionID, payload); err != nil {
		status := http.StatusBadRequest
		if err.Error() == "unauthorized" {
			status = http.StatusUnauthorized
		}
		c.JSON(status, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Password updated",
	})
}

// UpdateUserPermission 处理权限修改请求。
func (h *WebHandler) UpdateUserPermission(c *gin.Context) {
	sessionID, err := c.Cookie(session.CookieName)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "unauthorized",
		})
		return
	}

	var payload model.UserPermissionUpdate
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid request body",
		})
		return
	}

	if err := h.authService.UpdateUserPermission(sessionID, payload); err != nil {
		status := http.StatusBadRequest
		switch err.Error() {
		case "unauthorized":
			status = http.StatusUnauthorized
		case "only admin can update user permission":
			status = http.StatusForbidden
		case "user not found":
			status = http.StatusNotFound
		}

		c.JSON(status, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Permission updated",
	})
}

// ListUsers 返回后台用户列表。
func (h *WebHandler) ListUsers(c *gin.Context) {
	sessionID, err := c.Cookie(session.CookieName)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "unauthorized",
		})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	result, err := h.authService.ListUsers(sessionID, page, pageSize)
	if err != nil {
		status := http.StatusBadRequest
		switch err.Error() {
		case "unauthorized":
			status = http.StatusUnauthorized
		case "forbidden":
			status = http.StatusForbidden
		}

		c.JSON(status, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, userListResponse{
		Items:    result.Items,
		Page:     result.Page,
		PageSize: result.PageSize,
		Total:    result.Total,
	})
}

// DeleteUser 处理用户删除请求。
func (h *WebHandler) DeleteUser(c *gin.Context) {
	sessionID, err := c.Cookie(session.CookieName)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "unauthorized",
		})
		return
	}

	if err := h.authService.DeleteUser(sessionID, c.Param("username")); err != nil {
		status := http.StatusBadRequest
		switch err.Error() {
		case "unauthorized":
			status = http.StatusUnauthorized
		case "forbidden":
			status = http.StatusForbidden
		case "user not found":
			status = http.StatusNotFound
		case "cannot delete admin user":
			status = http.StatusForbidden
		case "user_admin can only delete user":
			status = http.StatusForbidden
		case "cannot delete current user":
			status = http.StatusBadRequest
		}

		c.JSON(status, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User deleted",
	})
}

// ReviewBlog 处理博客审核请求。
func (h *WebHandler) ReviewBlog(c *gin.Context) {
	user := h.getCurrentUser(c)
	if !user.IsLogin {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "unauthorized",
		})
		return
	}

	blogID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || blogID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid blog id",
		})
		return
	}

	var payload blogReviewRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid request body",
		})
		return
	}

	if err := h.blogService.ReviewBlog(blogID, payload.Status, payload.IsTop, user.Permission); err != nil {
		statusCode := http.StatusBadRequest
		switch err.Error() {
		case "unauthorized":
			statusCode = http.StatusUnauthorized
		case "forbidden":
			statusCode = http.StatusForbidden
		case "blog not found":
			statusCode = http.StatusNotFound
		}

		c.JSON(statusCode, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Blog reviewed",
	})
}

// SubmitGet 处理 submit 的 GET 请求。
func (h *WebHandler) SubmitGet(c *gin.Context) {
	c.JSON(http.StatusBadRequest, gin.H{
		"400": "Access method denied",
	})
}

// SubmitPost 处理 submit 的 POST 请求。
func (h *WebHandler) SubmitPost(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "POST test",
	})
}

// getCurrentUser 根据 Cookie 读取当前用户。
func (h *WebHandler) getCurrentUser(c *gin.Context) model.UserView {
	sessionID, err := c.Cookie(session.CookieName)
	if err != nil {
		return model.UserView{}
	}

	user, err := h.authService.CurrentUser(sessionID)
	if err != nil {
		return model.UserView{}
	}

	return user
}

// saveAvatarFile 校验并保存头像文件。
func splitTagInput(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}

	fields := strings.FieldsFunc(raw, func(r rune) bool {
		return r == ',' || r == '，' || r == ';' || r == '；' || r == '\n'
	})

	result := make([]string, 0, len(fields))
	for _, field := range fields {
		field = strings.TrimSpace(field)
		if field == "" {
			continue
		}
		result = append(result, field)
	}

	return result
}

func saveAvatarFile(fileHeader *multipart.FileHeader) (string, error) {
	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	allowedContentType, ok := allowedAvatarTypes[ext]
	if !ok {
		return "", errors.New("only png jpg jpeg gif images are allowed")
	}

	file, err := fileHeader.Open()
	if err != nil {
		return "", errors.New("failed to open uploaded file")
	}
	defer file.Close()

	head := make([]byte, 512)
	n, err := file.Read(head)
	if err != nil && !errors.Is(err, io.EOF) {
		return "", errors.New("failed to read uploaded file")
	}

	contentType := http.DetectContentType(head[:n])
	if contentType != allowedContentType {
		return "", errors.New("file type does not match the allowed image format")
	}

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return "", errors.New("failed to reset uploaded file stream")
	}

	dir := filepath.Join("frontend", "img")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", errors.New("failed to prepare avatar directory")
	}

	fileName, err := generateAvatarFileName(ext)
	if err != nil {
		return "", errors.New("failed to generate avatar file name")
	}

	targetPath := filepath.Join(dir, fileName)
	targetFile, err := os.Create(targetPath)
	if err != nil {
		return "", errors.New("failed to create avatar file")
	}
	defer targetFile.Close()

	if _, err := io.Copy(targetFile, file); err != nil {
		return "", errors.New("failed to save avatar file")
	}

	return fileName, nil
}

// generateAvatarFileName 生成头像文件名。
func generateAvatarFileName(ext string) (string, error) {
	randomBytes := make([]byte, 32)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}

	hash := sha256.Sum256(randomBytes)
	return hex.EncodeToString(hash[:]) + ext, nil
}

// GetBlogByID 返回博客详情。
func (h *WebHandler) GetBlogByID(c *gin.Context) {
	blogID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || blogID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid blog id",
		})
		return
	}

	user := h.getCurrentUser(c)
	blog, err := h.blogService.GetBlogByIDForUser(blogID, user.UserName, user.Permission)
	if err != nil {
		status := http.StatusInternalServerError
		switch err.Error() {
		case "blog not found":
			status = http.StatusNotFound
		case "forbidden":
			status = http.StatusForbidden
		}

		c.JSON(status, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, blog)
}

// CreateBlog 处理博客创建请求。
func (h *WebHandler) ToggleBlogLike(c *gin.Context) {
	user := h.getCurrentUser(c)
	blogID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || blogID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid blog id",
		})
		return
	}

	result, err := h.blogService.ToggleLike(blogID, user.UserName, user.Permission)
	if err != nil {
		statusCode := http.StatusBadRequest
		switch err.Error() {
		case "unauthorized":
			statusCode = http.StatusUnauthorized
		case "blog not found":
			statusCode = http.StatusNotFound
		case "forbidden":
			statusCode = http.StatusForbidden
		}

		c.JSON(statusCode, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, blogInteractionResponse{
		Active:        result.Active,
		LikeCount:     result.LikeCount,
		FavoriteCount: result.FavoriteCount,
	})
}

func (h *WebHandler) ToggleBlogFavorite(c *gin.Context) {
	user := h.getCurrentUser(c)
	blogID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || blogID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid blog id",
		})
		return
	}

	result, err := h.blogService.ToggleFavorite(blogID, user.UserName, user.Permission)
	if err != nil {
		statusCode := http.StatusBadRequest
		switch err.Error() {
		case "unauthorized":
			statusCode = http.StatusUnauthorized
		case "blog not found":
			statusCode = http.StatusNotFound
		case "forbidden":
			statusCode = http.StatusForbidden
		}

		c.JSON(statusCode, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, blogInteractionResponse{
		Active:        result.Active,
		LikeCount:     result.LikeCount,
		FavoriteCount: result.FavoriteCount,
	})
}

func (h *WebHandler) CreateBlog(c *gin.Context) {
	user := h.getCurrentUser(c)
	if !user.IsLogin {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "unauthorized",
		})
		return
	}

	title := c.PostForm("title")
	content := c.PostForm("content")
	status := c.PostForm("status")
	isTop, _ := strconv.ParseBool(c.DefaultPostForm("isTop", "false"))
	categoryID, _ := strconv.ParseInt(c.DefaultPostForm("categoryId", "0"), 10, 64)
	tags := splitTagInput(c.PostForm("tags"))

	blog, err := h.blogService.CreateBlog(model.BlogCreateInput{
		Title:          title,
		Content:        content,
		Status:         status,
		IsTop:          isTop,
		CategoryID:     categoryID,
		Tags:           tags,
		AuthorUsername: user.UserName,
		Permission:     user.Permission,
	})
	if err != nil {
		statusCode := http.StatusBadRequest
		if err.Error() == "unauthorized" {
			statusCode = http.StatusUnauthorized
		}

		c.JSON(statusCode, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Blog created",
		"id":      blog.ID,
		"status":  blog.Status,
	})
}

// UpdateBlog 处理博客编辑请求。
func (h *WebHandler) UpdateBlog(c *gin.Context) {
	user := h.getCurrentUser(c)
	if !user.IsLogin {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "unauthorized",
		})
		return
	}

	blogID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || blogID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid blog id",
		})
		return
	}

	title := c.PostForm("title")
	content := c.PostForm("content")
	status := c.PostForm("status")
	isTop, _ := strconv.ParseBool(c.DefaultPostForm("isTop", "false"))
	categoryID, _ := strconv.ParseInt(c.DefaultPostForm("categoryId", "0"), 10, 64)
	tags := splitTagInput(c.PostForm("tags"))

	if err := h.blogService.UpdateBlog(model.BlogUpdateInput{
		BlogID:      blogID,
		Title:       title,
		Content:     content,
		Status:      status,
		IsTop:       isTop,
		CategoryID:  categoryID,
		Tags:        tags,
		CurrentUser: user.UserName,
		CurrentPerm: user.Permission,
	}); err != nil {
		statusCode := http.StatusBadRequest
		switch err.Error() {
		case "unauthorized":
			statusCode = http.StatusUnauthorized
		case "only the author can edit this blog":
			statusCode = http.StatusForbidden
		case "blog not found":
			statusCode = http.StatusNotFound
		}

		c.JSON(statusCode, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Blog updated",
	})
}

// DeleteBlog 处理博客删除请求。
func (h *WebHandler) DeleteBlog(c *gin.Context) {
	user := h.getCurrentUser(c)
	if !user.IsLogin {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "unauthorized",
		})
		return
	}

	blogID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || blogID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid blog id",
		})
		return
	}

	if err := h.blogService.DeleteBlog(blogID, user.UserName, user.Permission); err != nil {
		status := http.StatusBadRequest
		switch err.Error() {
		case "unauthorized":
			status = http.StatusUnauthorized
		case "only the author can delete this blog":
			status = http.StatusForbidden
		case "blog not found":
			status = http.StatusNotFound
		}

		c.JSON(status, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Blog deleted",
	})
}

// ListComments 返回博客评论列表。
func (h *WebHandler) ListComments(c *gin.Context) {
	blogID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || blogID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid blog id",
		})
		return
	}

	user := h.getCurrentUser(c)
	comments, err := h.commentService.ListComments(blogID, user.UserName, user.Permission)
	if err != nil {
		statusCode := http.StatusInternalServerError
		switch err.Error() {
		case "blog not found":
			statusCode = http.StatusNotFound
		case "forbidden":
			statusCode = http.StatusForbidden
		}

		c.JSON(statusCode, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items": comments,
	})
}

// CreateComment 创建博客评论。
func (h *WebHandler) CreateComment(c *gin.Context) {
	user := h.getCurrentUser(c)
	if !user.IsLogin {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "unauthorized",
		})
		return
	}

	blogID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || blogID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid blog id",
		})
		return
	}

	var payload commentCreateRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid request body",
		})
		return
	}

	comment, err := h.commentService.CreateComment(blogID, payload.Content, user.UserName, user.Permission)
	if err != nil {
		statusCode := http.StatusBadRequest
		switch err.Error() {
		case "unauthorized":
			statusCode = http.StatusUnauthorized
		case "blog not found":
			statusCode = http.StatusNotFound
		case "forbidden":
			statusCode = http.StatusForbidden
		}

		c.JSON(statusCode, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Comment created",
		"item":    comment,
	})
}

// DeleteComment 删除评论。
func (h *WebHandler) DeleteComment(c *gin.Context) {
	user := h.getCurrentUser(c)
	if !user.IsLogin {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "unauthorized",
		})
		return
	}

	commentID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || commentID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid comment id",
		})
		return
	}

	if err := h.commentService.DeleteComment(commentID, user.UserName, user.Permission); err != nil {
		statusCode := http.StatusBadRequest
		switch err.Error() {
		case "unauthorized":
			statusCode = http.StatusUnauthorized
		case "forbidden":
			statusCode = http.StatusForbidden
		case "comment not found":
			statusCode = http.StatusNotFound
		}

		c.JSON(statusCode, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Comment deleted",
	})
}
