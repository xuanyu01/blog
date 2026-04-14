/*
这个文件实现登录校验中间件
*/
package middleware

import (
	"blog/session"
	"net/http"

	"github.com/gin-gonic/gin"
)

// RequireLogin 创建要求用户已登录的中间件
func RequireLogin(sessionStore session.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionID, err := c.Cookie(session.CookieName)
		if err != nil {
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}

		userID, err := sessionStore.Get(sessionID)
		if err != nil || userID == "" {
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}

		c.Set("userID", userID)
		c.Next()
	}
}
