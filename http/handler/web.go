/*
web.go 提供前端页面依赖的公共 HTTP 处理逻辑和辅助函数。
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
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

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
	loginLimiter   loginRateLimiter
}

// loginRateLimiter 定义登录限流器需要提供的能力。
type loginRateLimiter interface {
	Check(key string) (time.Duration, error)
	RegisterFailure(key string) (time.Duration, error)
	Reset(key string) error
}

type noopLoginRateLimiter struct{}

func (noopLoginRateLimiter) Check(key string) (time.Duration, error)           { return 0, nil }
func (noopLoginRateLimiter) RegisterFailure(key string) (time.Duration, error) { return 0, nil }
func (noopLoginRateLimiter) Reset(key string) error                            { return nil }

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

// taxonomyListResponse 表示分类、标签和归档列表的通用响应结构。
type taxonomyListResponse[T any] struct {
	Items []T `json:"items"`
}

// blogReviewRequest 表示后台审核博客的请求体。
type blogReviewRequest struct {
	Status string `json:"status"`
	IsTop  bool   `json:"isTop"`
}

// blogInteractionResponse 表示点赞或收藏后的交互结果。
type blogInteractionResponse struct {
	Active        bool  `json:"active"`
	LikeCount     int64 `json:"likeCount"`
	FavoriteCount int64 `json:"favoriteCount"`
}

// categoryPayload 表示分类创建或更新的请求体。
type categoryPayload struct {
	Name string `json:"name"`
}

// commentCreateRequest 表示评论创建请求体。
type commentCreateRequest struct {
	Content string `json:"content"`
}

// NewWebHandler 创建处理器实例。
func NewWebHandler(blogService *service.BlogService, commentService *service.CommentService, authService *service.AuthService, loginLimiter loginRateLimiter) *WebHandler {
	if loginLimiter == nil {
		loginLimiter = noopLoginRateLimiter{}
	}

	return &WebHandler{
		blogService:    blogService,
		commentService: commentService,
		authService:    authService,
		loginLimiter:   loginLimiter,
	}
}

// GetAppState 返回首页所需的聚合状态。
func (h *WebHandler) GetAppState(c *gin.Context) {
	user := h.getCurrentUser(c)
	c.JSON(http.StatusOK, appStateResponse{
		User: user,
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

// splitTagInput 按中英文分隔符拆分标签输入。
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

// saveAvatarFile 校验头像文件并保存到本地目录。
func saveAvatarFile(fileHeader *multipart.FileHeader) (string, error) {
	// 先校验扩展名和文件头，再落盘保存文件。
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

// generateAvatarFileName 生成随机且稳定长度的头像文件名。
func generateAvatarFileName(ext string) (string, error) {
	randomBytes := make([]byte, 32)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}

	hash := sha256.Sum256(randomBytes)
	return hex.EncodeToString(hash[:]) + ext, nil
}
