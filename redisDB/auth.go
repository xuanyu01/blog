package redisDB

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// 登录验证中间件
func AuthMiddleware() gin.HandlerFunc {

	return func(c *gin.Context) {

		// 从cookie获取session_id
		sessionID, err := c.Cookie("session_id")
		if err != nil {
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}

		// 查询Redis
		userID, err := GetSession(sessionID)
		if err != nil || userID == "" {
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}

		// 保存用户ID到context
		c.Set("userID", userID)

		c.Next()
	}
}
