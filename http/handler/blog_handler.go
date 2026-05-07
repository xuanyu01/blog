/*
处理博客、收藏、互动和后台博客管理相关接口。
*/
package handler

import (
	"blog/model"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

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

// ListLikedBlogs 返回当前用户点赞过的博客列表。
func (h *WebHandler) ListLikedBlogs(c *gin.Context) {
	user := h.getCurrentUser(c)
	if !user.IsLogin {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "unauthorized",
		})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	result, err := h.blogService.ListLikedBlogs(page, pageSize, user.UserName)
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

// ToggleBlogLike 处理点赞切换请求。
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

// ToggleBlogFavorite 处理收藏切换请求。
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

// CreateBlog 处理博客创建请求。
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
