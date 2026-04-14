/*
该文件实现前端页面依赖的 HTTP 处理逻辑
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
	"strings"

	"github.com/gin-gonic/gin"
)

var allowedAvatarTypes = map[string]string{
	".png":  "image/png",
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".gif":  "image/gif",
}

// WebHandler 负责处理博客和认证相关请求
// 它位于 HTTP 层 把请求参数转换为服务层调用
type WebHandler struct {
	blogService *service.BlogService
	authService *service.AuthService
}

// appStateResponse 表示首页状态接口的返回结构
type appStateResponse struct {
	Blogs []model.Blog   `json:"blogs"`
	User  model.UserView `json:"user"`
}

// NewWebHandler 创建 WebHandler 实例
func NewWebHandler(blogService *service.BlogService, authService *service.AuthService) *WebHandler {
	return &WebHandler{
		blogService: blogService,
		authService: authService,
	}
}

// GetAppState 返回首页所需的聚合数据
func (h *WebHandler) GetAppState(c *gin.Context) {
	user := h.getCurrentUser(c)
	blogs, err := h.blogService.ListBlogs()
	if err != nil {
		log.Println("list blogs failed:", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to load blogs",
		})
		return
	}

	c.JSON(http.StatusOK, appStateResponse{
		Blogs: blogs,
		User:  user,
	})
}

// Register 处理注册请求
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

// Login 处理登录请求
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

// Logout 处理退出登录请求
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

// CurrentUser 返回当前登录用户信息
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

// UpdateProfile 处理用户资料修改请求
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

// UploadAvatar 处理头像上传请求
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

// UpdatePassword 处理用户密码修改请求
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

// SubmitGet 处理 submit 的 GET 请求
func (h *WebHandler) SubmitGet(c *gin.Context) {
	c.JSON(http.StatusBadRequest, gin.H{
		"400": "Access method denied",
	})
}

// SubmitPost 处理 submit 的 POST 请求
func (h *WebHandler) SubmitPost(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "POST test",
	})
}

// getCurrentUser 根据 Cookie 解析当前用户状态
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

// saveAvatarFile 校验并保存头像文件
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

// generateAvatarFileName 生成不可逆且难以猜测的头像文件名
func generateAvatarFileName(ext string) (string, error) {
	randomBytes := make([]byte, 32)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}

	hash := sha256.Sum256(randomBytes)
	return hex.EncodeToString(hash[:]) + ext, nil
}
