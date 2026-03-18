package http

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"youmind-backend-v2/internal/application/auth"
	"youmind-backend-v2/internal/domain/user"
)

const currentUserKey = "currentUser"

// AuthMiddleware 从 Bearer token 解析用户并注入 context
func AuthMiddleware(authSvc *auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"detail": "Missing Authorization header"})
			return
		}
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"detail": "Invalid Authorization header"})
			return
		}
		parsed, err := jwt.Parse(parts[1], func(token *jwt.Token) (interface{}, error) {
			return []byte(authSvc.Secret()), nil
		})
		if err != nil || !parsed.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"detail": "Could not validate credentials"})
			return
		}
		claims, ok := parsed.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"detail": "Could not validate credentials"})
			return
		}
		sub, _ := claims["sub"].(string)
		if sub == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"detail": "Could not validate credentials"})
			return
		}
		u, err := authSvc.GetUserByID(sub)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"detail": "Could not validate credentials"})
			return
		}
		c.Set(currentUserKey, u)
		c.Next()
	}
}

// GetCurrentUser 从 context 取出当前用户（仅中间件之后使用）
func GetCurrentUser(c *gin.Context) (*user.User, bool) {
	v, ok := c.Get(currentUserKey)
	if !ok {
		return nil, false
	}
	u, ok := v.(*user.User)
	return u, ok
}
