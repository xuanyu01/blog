/*
auth.go 提供登录校验中间件。*/
package middleware

import (
	"blog/model"
	"blog/session"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

// currentUserProvider 定义读取当前用户信息所需的能力。
type currentUserProvider interface {
	CurrentUser(sessionID string) (model.UserView, error)
}

// RequireLogin 创建要求用户已登录的中间件。
func RequireLogin(sessionStore session.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionID, userID, err := readSession(c, sessionStore)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"message": "unauthorized",
			})
			c.Abort()
			return
		}

		c.Set("sessionID", sessionID)
		c.Set("userID", userID)
		c.Next()
	}
}

// RequireManager 。。。。Ҫ。。。û。ӵ。й。。。Ȩ。޵。。м。。。。
func RequireManager(sessionStore session.Store, userProvider currentUserProvider) gin.HandlerFunc {
	return requirePermission(sessionStore, userProvider, model.CanManageAllBlogs)
}

// RequireAdmin 。。。。Ҫ。。。û。ӵ。й。。。ԱȨ。޵。。м。。。。
func RequireAdmin(sessionStore session.Store, userProvider currentUserProvider) gin.HandlerFunc {
	return requirePermission(sessionStore, userProvider, func(permission string) bool {
		return permission == model.PermissionAdmin
	})
}

// requirePermission 校验当前用户是否满足指定权限条件。
func requirePermission(sessionStore session.Store, userProvider currentUserProvider, allow func(string) bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionID, userID, err := readSession(c, sessionStore)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"message": "unauthorized",
			})
			c.Abort()
			return
		}

		c.Set("sessionID", sessionID)
		c.Set("userID", userID)

		user, err := userProvider.CurrentUser(sessionID)
		if err != nil || !user.IsLogin {
			c.JSON(http.StatusUnauthorized, gin.H{
				"message": "unauthorized",
			})
			c.Abort()
			return
		}
		if !allow(user.Permission) {
			c.JSON(http.StatusForbidden, gin.H{
				"message": "forbidden",
			})
			c.Abort()
			return
		}

		c.Set("currentUser", user)
		c.Next()
	}
}

// readSession 从会话存储中读取当前请求对应的会话信息。
func readSession(c *gin.Context, sessionStore session.Store) (string, string, error) {
	sessionID, err := c.Cookie(session.CookieName)
	if err != nil {
		return "", "", err
	}

	userID, err := sessionStore.Get(sessionID)
	if err != nil || userID == "" {
		if err == nil {
			err = errors.New("session not found")
		}
		return "", "", err
	}

	return sessionID, userID, nil
}

