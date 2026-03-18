package http

import (
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"youmind-backend-v2/internal/application/auth"
	"youmind-backend-v2/internal/config"
	"youmind-backend-v2/internal/domain/user"
)

// AuthHandler 认证相关 HTTP 处理
type AuthHandler struct {
	svc *auth.Service
	cfg *config.Config
}

func NewAuthHandler(svc *auth.Service, cfg *config.Config) *AuthHandler {
	return &AuthHandler{svc: svc, cfg: cfg}
}

func (h *AuthHandler) RegisterRoutes(r *gin.RouterGroup) {
	r.POST("/register", h.register)
	r.POST("/login", h.login)
	authGroup := r.Group("")
	authGroup.Use(AuthMiddleware(h.svc))
	authGroup.GET("/me", h.me)
	authGroup.PATCH("/me", h.updateMe)
}

type registerRequest struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type loginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type userResponse struct {
	ID               string `json:"id"`
	Username         string `json:"username"`
	Email            string `json:"email"`
	SubscriptionPlan string `json:"subscription_plan"`
	CreditsBalance   int    `json:"credits_balance"`
	CreditsUsed      int    `json:"credits_used"`
	CreatedAt        string `json:"created_at"`
}

type updateMeRequest struct {
	Username *string `json:"username,omitempty"`
	Email    *string `json:"email,omitempty"`
}

func toUserResponse(u *user.User) userResponse {
	return userResponse{
		ID:               u.ID,
		Username:         u.Username,
		Email:            u.Email,
		SubscriptionPlan: u.SubscriptionPlan,
		CreditsBalance:   u.CreditsBalance,
		CreditsUsed:      u.CreditsUsed,
		CreatedAt:        u.CreatedAt.Format("2006-01-02T15:04:05.000Z"),
	}
}

func (h *AuthHandler) register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "invalid request body"})
		return
	}
	result, err := h.svc.Register(auth.RegisterInput{
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		if err == auth.ErrUsernameOrEmailTaken {
			c.JSON(http.StatusBadRequest, gin.H{"detail": "Username or email already registered"})
			return
		}
		log.Printf("register: %v", err)
		detail := "failed to create user"
		if h.cfg.AppEnv == "development" {
			detail = "failed to create user: " + err.Error()
		}
		c.JSON(http.StatusInternalServerError, gin.H{"detail": detail})
		return
	}
	c.JSON(http.StatusOK, userResponse{
		ID:               result.ID,
		Username:         result.Username,
		Email:            result.Email,
		SubscriptionPlan: result.SubscriptionPlan,
		CreditsBalance:   result.CreditsBalance,
		CreditsUsed:      result.CreditsUsed,
		CreatedAt:        result.CreatedAt.Format("2006-01-02T15:04:05.000Z"),
	})
}

func (h *AuthHandler) login(c *gin.Context) {
	var req loginRequest
	// 先按 Content-Type 解析，避免 ShouldBindJSON 消费 body 导致后续 PostForm 读不到
	if strings.Contains(c.GetHeader("Content-Type"), "application/json") {
		_ = c.ShouldBindJSON(&req)
	}
	if req.Username == "" {
		req.Username = c.PostForm("username")
	}
	if req.Password == "" {
		req.Password = c.PostForm("password")
	}
	if req.Username == "" || req.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "username and password required"})
		return
	}
	result, err := h.svc.Login(auth.LoginInput{Username: req.Username, Password: req.Password})
	if err != nil {
		if err == auth.ErrInvalidCredentials {
			c.JSON(http.StatusUnauthorized, gin.H{"detail": "Incorrect username or password"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "login failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"access_token": result.AccessToken,
		"token_type":   result.TokenType,
	})
}

func (h *AuthHandler) me(c *gin.Context) {
	u, ok := GetCurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"detail": "Not authenticated"})
		return
	}
	c.JSON(http.StatusOK, toUserResponse(u))
}

func (h *AuthHandler) updateMe(c *gin.Context) {
	u, ok := GetCurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"detail": "Not authenticated"})
		return
	}
	var req updateMeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "invalid request body"})
		return
	}
	updated, err := h.svc.UpdateProfile(u.ID, req.Username, req.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "failed to update user"})
		return
	}
	c.JSON(http.StatusOK, toUserResponse(updated))
}
