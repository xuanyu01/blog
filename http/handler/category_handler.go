/*
category_handler.go 。。。。。。ࡢ。。ǩ。͹鵵。。ؽӿڡ。
*/
package handler

import (
	"blog/model"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ListCategories 返回分类列表。
func (h *WebHandler) ListCategories(c *gin.Context) {
	items, err := h.blogService.ListCategories()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to load categories",
		})
		return
	}

	c.JSON(http.StatusOK, taxonomyListResponse[model.Category]{
		Items: items,
	})
}

// ListTags 返回标签列表。
func (h *WebHandler) ListTags(c *gin.Context) {
	items, err := h.blogService.ListTags()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to load tags",
		})
		return
	}

	c.JSON(http.StatusOK, taxonomyListResponse[model.Tag]{
		Items: items,
	})
}

// ListArchives 返回归档列表。
func (h *WebHandler) ListArchives(c *gin.Context) {
	items, err := h.blogService.ListArchives()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to load archives",
		})
		return
	}

	c.JSON(http.StatusOK, taxonomyListResponse[model.ArchiveItem]{
		Items: items,
	})
}

// ListManageCategories 返回后台分类列表。
func (h *WebHandler) ListManageCategories(c *gin.Context) {
	user := h.getCurrentUser(c)
	if !user.IsLogin {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "unauthorized"})
		return
	}

	items, err := h.blogService.ListManageCategories(user.Permission)
	if err != nil {
		statusCode := http.StatusBadRequest
		if err.Error() == "forbidden" {
			statusCode = http.StatusForbidden
		}
		c.JSON(statusCode, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, taxonomyListResponse[model.Category]{Items: items})
}

// CreateCategory 创建分类。
func (h *WebHandler) CreateCategory(c *gin.Context) {
	user := h.getCurrentUser(c)
	if !user.IsLogin {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "unauthorized"})
		return
	}

	var payload categoryPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid request body"})
		return
	}

	category, err := h.blogService.CreateCategory(payload.Name, user.Permission)
	if err != nil {
		statusCode := http.StatusBadRequest
		if err.Error() == "forbidden" {
			statusCode = http.StatusForbidden
		}
		c.JSON(statusCode, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"item": category})
}

// UpdateCategory 更新分类。
func (h *WebHandler) UpdateCategory(c *gin.Context) {
	user := h.getCurrentUser(c)
	if !user.IsLogin {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "unauthorized"})
		return
	}

	categoryID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || categoryID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid category id"})
		return
	}

	var payload categoryPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid request body"})
		return
	}

	category, err := h.blogService.UpdateCategory(categoryID, payload.Name, user.Permission)
	if err != nil {
		statusCode := http.StatusBadRequest
		switch err.Error() {
		case "forbidden":
			statusCode = http.StatusForbidden
		case "category not found":
			statusCode = http.StatusNotFound
		}
		c.JSON(statusCode, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"item": category})
}

// DeleteCategory 删除分类。
func (h *WebHandler) DeleteCategory(c *gin.Context) {
	user := h.getCurrentUser(c)
	if !user.IsLogin {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "unauthorized"})
		return
	}

	categoryID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || categoryID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid category id"})
		return
	}

	if err := h.blogService.DeleteCategory(categoryID, user.Permission); err != nil {
		statusCode := http.StatusBadRequest
		switch err.Error() {
		case "forbidden":
			statusCode = http.StatusForbidden
		case "category not found":
			statusCode = http.StatusNotFound
		}
		c.JSON(statusCode, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Category deleted"})
}

