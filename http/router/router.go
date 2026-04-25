/*
该文件负责注册 HTTP 路由并提供前端构建产物访问
*/
package router

import (
	"blog/http/handler"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

// New 创建并配置 Gin 路由引擎
func New(webHandler *handler.WebHandler) *gin.Engine {
	r := gin.Default()

	// 图片资源统一从 frontend/img 提供 这样用户上传的新头像可以直接被访问
	r.Static("/img", filepath.Join("frontend", "img"))

	api := r.Group("/api")
	api.GET("/state", webHandler.GetAppState)
	api.POST("/register", webHandler.Register)
	api.POST("/login", webHandler.Login)
	api.POST("/logout", webHandler.Logout)
	api.GET("/me", webHandler.CurrentUser)
	api.PUT("/user/profile", webHandler.UpdateProfile)
	api.PUT("/user/password", webHandler.UpdatePassword)
	api.PUT("/user/permission", webHandler.UpdateUserPermission)
	api.GET("/admin/users", webHandler.ListUsers)
	api.DELETE("/admin/users/:username", webHandler.DeleteUser)
	api.POST("/user/avatar", webHandler.UploadAvatar)
	api.GET("/submit", webHandler.SubmitGet)
	api.POST("/submit", webHandler.SubmitPost)
	api.GET("/blogs", webHandler.ListBlogs)
	api.POST("/blogs", webHandler.CreateBlog)
	api.GET("/blogs/:id", webHandler.GetBlogByID)
	api.PUT("/blogs/:id", webHandler.UpdateBlog)
	api.DELETE("/blogs/:id", webHandler.DeleteBlog)

	distDir := filepath.Join("frontend", "dist")
	assetsDir := filepath.Join(distDir, "assets")
	indexPath := filepath.Join(distDir, "index.html")

	// 只有当前端构建产物存在时才注册静态资源目录 避免启动时因为缺文件报错
	if stat, err := os.Stat(assetsDir); err == nil && stat.IsDir() {
		r.Static("/assets", assetsDir)
	}

	r.NoRoute(func(c *gin.Context) {
		// API 路由不存在时返回 JSON 前端路由不存在时交给 SPA 入口处理
		if strings.HasPrefix(c.Request.URL.Path, "/api/") {
			c.JSON(http.StatusNotFound, gin.H{
				"message": "not found",
			})
			return
		}

		if _, err := os.Stat(indexPath); err != nil {
			c.String(http.StatusServiceUnavailable, "frontend build not found, please run npm run build in frontend")
			return
		}

		c.File(indexPath)
	})

	return r
}
