package http

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// ModelsHandler 模型列表（占位，与旧版兼容）
type ModelsHandler struct{}

func NewModelsHandler() *ModelsHandler {
	return &ModelsHandler{}
}

func (h *ModelsHandler) RegisterRoutes(r *gin.RouterGroup) {
	r.GET("", h.listModels)
}

func (h *ModelsHandler) listModels(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"models":        []gin.H{},
		"default_model": "",
		"updated_at":    time.Now().UTC(),
	})
}
