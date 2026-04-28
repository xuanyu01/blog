/*
router.go 负责注册 HTTP 路由并提供前端静态资源访问。
*/
package router

import (
	"blog/http/handler"
	"blog/http/middleware"
	"blog/service"
	"blog/session"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

// New 创建并配置 Gin 路由。
func New(webHandler *handler.WebHandler, sessionStore session.Store, authService *service.AuthService) *gin.Engine {
	r := gin.Default()
	_ = r.SetTrustedProxies([]string{"127.0.0.1", "::1"})

	// 用户上传的头像统一从 frontend/img 提供访问。
	r.Static("/img", filepath.Join("frontend", "img"))

	api := r.Group("/api")
	api.GET("/state", webHandler.GetAppState)
	api.POST("/register", webHandler.Register)
	api.POST("/login", webHandler.Login)
	api.GET("/submit", webHandler.SubmitGet)
	api.POST("/submit", webHandler.SubmitPost)
	api.GET("/categories", webHandler.ListCategories)
	api.GET("/tags", webHandler.ListTags)
	api.GET("/archives", webHandler.ListArchives)
	api.GET("/blogs", webHandler.ListBlogs)
	api.GET("/blogs/:id", webHandler.GetBlogByID)
	api.GET("/blogs/:id/comments", webHandler.ListComments)

	authRequired := api.Group("")
	authRequired.Use(middleware.RequireLogin(sessionStore))
	authRequired.POST("/logout", webHandler.Logout)
	authRequired.GET("/me", webHandler.CurrentUser)
	authRequired.PUT("/user/profile", webHandler.UpdateProfile)
	authRequired.PUT("/user/password", webHandler.UpdatePassword)
	authRequired.GET("/user/blogs", webHandler.ListCurrentUserBlogs)
	authRequired.GET("/user/favorites", webHandler.ListFavoriteBlogs)
	authRequired.POST("/user/avatar", webHandler.UploadAvatar)
	authRequired.POST("/blogs", webHandler.CreateBlog)
	authRequired.POST("/blogs/:id/like", webHandler.ToggleBlogLike)
	authRequired.POST("/blogs/:id/favorite", webHandler.ToggleBlogFavorite)
	authRequired.POST("/blogs/:id/comments", webHandler.CreateComment)
	authRequired.PUT("/blogs/:id", webHandler.UpdateBlog)
	authRequired.DELETE("/blogs/:id", webHandler.DeleteBlog)
	authRequired.DELETE("/comments/:id", webHandler.DeleteComment)

	managerRoutes := api.Group("")
	managerRoutes.Use(middleware.RequireManager(sessionStore, authService))
	managerRoutes.GET("/admin/users", webHandler.ListUsers)
	managerRoutes.DELETE("/admin/users/:username", webHandler.DeleteUser)
	managerRoutes.GET("/admin/blogs", webHandler.ListManagedBlogs)
	managerRoutes.GET("/admin/categories", webHandler.ListManageCategories)
	managerRoutes.POST("/admin/categories", webHandler.CreateCategory)
	managerRoutes.PUT("/admin/categories/:id", webHandler.UpdateCategory)
	managerRoutes.DELETE("/admin/categories/:id", webHandler.DeleteCategory)
	managerRoutes.PUT("/admin/blogs/:id/review", webHandler.ReviewBlog)

	adminRoutes := api.Group("")
	adminRoutes.Use(middleware.RequireAdmin(sessionStore, authService))
	adminRoutes.PUT("/user/permission", webHandler.UpdateUserPermission)

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
