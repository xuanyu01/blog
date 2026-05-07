/*
处理注册、登录、登出和用户资料相关接口。
*/
package handler

import (
	"blog/model"
	"blog/service"
	"blog/session"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

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
	username := strings.TrimSpace(c.PostForm("username"))
	password := c.PostForm("password")
	loginLimitKey := buildLoginLimitKey(c.ClientIP(), username)

	// 登录限流检查
	if retryAfter, err := h.loginLimiter.Check(loginLimitKey); err != nil {
		log.Println("login rate limit check failed:", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "login is temporarily unavailable",
		})
		return
	} else if retryAfter > 0 {
		c.JSON(http.StatusTooManyRequests, gin.H{
			"message":    "too many login attempts, please try again later",
			"retryAfter": int(retryAfter.Seconds()),
		})
		return
	}

	// 校验用户名和密码并创建会话
	sessionID, err := h.authService.Login(username, password)
	if err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, service.ErrInvalidCredentials) {
			retryAfter, limitErr := h.loginLimiter.RegisterFailure(loginLimitKey)
			if limitErr != nil {
				log.Println("login rate limit register failed:", limitErr)
			}
			if retryAfter > 0 {
				status = http.StatusTooManyRequests
				c.JSON(status, gin.H{
					"message":    "too many login attempts, please try again later",
					"retryAfter": int(retryAfter.Seconds()),
				})
				return
			}
		} else {
			status = http.StatusInternalServerError
			log.Println("login failed:", err)
		}

		c.JSON(status, gin.H{
			"message": err.Error(),
		})
		return
	}

	if err := h.loginLimiter.Reset(loginLimitKey); err != nil {
		log.Println("login rate limit reset failed:", err)
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

	// 获取oldUser信息
	oldUser, err := h.authService.CurrentUser(sessionID)
	if err != nil || !oldUser.IsLogin {
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

	// 更新用户头像
	user, err := h.authService.UpdateAvatar(sessionID, fileName)
	if err != nil {
		if deleteErr := deleteAvatarFile(fileName); deleteErr != nil {
			log.Println("delete new avatar after update failed:", deleteErr)
		}
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	// 删除旧头像文件（如果有且已更换）
	if oldUser.ImageRoute != "" && oldUser.ImageRoute != user.ImageRoute {
		if deleteErr := deleteAvatarFile(oldUser.ImageRoute); deleteErr != nil {
			log.Println("delete old avatar failed:", deleteErr)
		}
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

	// 解析请求体
	var payload model.UserPermissionUpdate
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid request body",
		})
		return
	}

	// 更新用户权限
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

// buildLoginLimitKey 根据 IP 和用户名生成登录限流键。
func buildLoginLimitKey(ip, username string) string {
	raw := strings.TrimSpace(ip) + "|" + strings.ToLower(strings.TrimSpace(username))
	if raw == "|" {
		raw = "anonymous"
	}

	hash := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(hash[:])
}
