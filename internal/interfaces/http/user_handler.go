package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"youmind-backend-v2/internal/application/user"
)

// UserHandler 用户相关 HTTP（余额等）
type UserHandler struct {
	svc *user.Service
}

func NewUserHandler(svc *user.Service) *UserHandler {
	return &UserHandler{svc: svc}
}

func (h *UserHandler) RegisterRoutes(r *gin.RouterGroup) {
	r.GET("/balance", h.balance)
}

func (h *UserHandler) balance(c *gin.Context) {
	u, ok := GetCurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"detail": "Not authenticated"})
		return
	}
	result, err := h.svc.GetBalance(u.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "failed to load balance"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"balance": result.Balance,
		"used":    result.Used,
		"plan":    result.Plan,
	})
}
