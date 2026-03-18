package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Health 健康检查
func Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "healthy"})
}
