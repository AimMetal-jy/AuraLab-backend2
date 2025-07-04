package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// TestHandler 测试处理器
func TestHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Test handler works",
		"success": true,
	})
}
