/*
注册 HTTP 路由并提供前端资源访问。
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

// New 创建并配置 Gin 路由。
func New(webHandler *handler.WebHandler) *gin.Engine {
	r := gin.Default()
	_ = r.SetTrustedProxies([]string{"127.0.0.1", "::1"})

	// 用户上传的头像统一从 frontend/img 提供访问。
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
	api.GET("/user/blogs", webHandler.ListCurrentUserBlogs)
	api.GET("/user/favorites", webHandler.ListFavoriteBlogs)
	api.GET("/admin/users", webHandler.ListUsers)
	api.DELETE("/admin/users/:username", webHandler.DeleteUser)
	api.GET("/admin/blogs", webHandler.ListManagedBlogs)
	api.GET("/admin/categories", webHandler.ListManageCategories)
	api.POST("/admin/categories", webHandler.CreateCategory)
	api.PUT("/admin/categories/:id", webHandler.UpdateCategory)
	api.DELETE("/admin/categories/:id", webHandler.DeleteCategory)
	api.PUT("/admin/blogs/:id/review", webHandler.ReviewBlog)
	api.POST("/user/avatar", webHandler.UploadAvatar)
	api.GET("/submit", webHandler.SubmitGet)
	api.POST("/submit", webHandler.SubmitPost)
	api.GET("/categories", webHandler.ListCategories)
	api.GET("/tags", webHandler.ListTags)
	api.GET("/archives", webHandler.ListArchives)
	api.GET("/blogs", webHandler.ListBlogs)
	api.POST("/blogs", webHandler.CreateBlog)
	api.GET("/blogs/:id", webHandler.GetBlogByID)
	api.POST("/blogs/:id/like", webHandler.ToggleBlogLike)
	api.POST("/blogs/:id/favorite", webHandler.ToggleBlogFavorite)
	api.GET("/blogs/:id/comments", webHandler.ListComments)
	api.POST("/blogs/:id/comments", webHandler.CreateComment)
	api.PUT("/blogs/:id", webHandler.UpdateBlog)
	api.DELETE("/blogs/:id", webHandler.DeleteBlog)
	api.DELETE("/comments/:id", webHandler.DeleteComment)

	distDir := filepath.Join("frontend", "dist")
	assetsDir := filepath.Join(distDir, "assets")
	indexPath := filepath.Join(distDir, "index.html")

	// 只有构建产物存在时才注册静态资源目录。
	if stat, err := os.Stat(assetsDir); err == nil && stat.IsDir() {
		r.Static("/assets", assetsDir)
	}

	r.NoRoute(func(c *gin.Context) {
		// API 未命中时返回 JSON，其余请求交给前端入口文件。
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
