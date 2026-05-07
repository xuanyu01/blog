/*
处理评论查询、创建和删除接口。
*/
package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ListComments 返回博客评论列表。
func (h *WebHandler) ListComments(c *gin.Context) {
	blogID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || blogID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid blog id",
		})
		return
	}

	user := h.getCurrentUser(c)
	comments, err := h.commentService.ListComments(blogID, user.UserName, user.Permission)
	if err != nil {
		statusCode := http.StatusInternalServerError
		switch err.Error() {
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

	c.JSON(http.StatusOK, gin.H{
		"items": comments,
	})
}

// CreateComment 创建博客评论或回复。
func (h *WebHandler) CreateComment(c *gin.Context) {
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

	var payload commentCreateRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid request body",
		})
		return
	}

	comment, err := h.commentService.CreateComment(blogID, payload.ParentID, payload.Content, user.UserName, user.Permission)
	if err != nil {
		statusCode := http.StatusBadRequest
		switch err.Error() {
		case "unauthorized":
			statusCode = http.StatusUnauthorized
		case "blog not found", "parent comment not found":
			statusCode = http.StatusNotFound
		case "forbidden":
			statusCode = http.StatusForbidden
		}

		c.JSON(statusCode, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Comment created",
		"item":    comment,
	})
}

// DeleteComment 删除评论。
func (h *WebHandler) DeleteComment(c *gin.Context) {
	user := h.getCurrentUser(c)
	if !user.IsLogin {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "unauthorized",
		})
		return
	}

	commentID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || commentID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid comment id",
		})
		return
	}

	if err := h.commentService.DeleteComment(commentID, user.UserName, user.Permission); err != nil {
		statusCode := http.StatusBadRequest
		switch err.Error() {
		case "unauthorized":
			statusCode = http.StatusUnauthorized
		case "forbidden":
			statusCode = http.StatusForbidden
		case "comment not found":
			statusCode = http.StatusNotFound
		}

		c.JSON(statusCode, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Comment deleted",
	})
}
