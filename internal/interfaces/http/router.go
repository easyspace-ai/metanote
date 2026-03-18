package http

import (
	"net/http"
	"path/filepath"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"youmind-backend-v2/internal/application/auth"
	"youmind-backend-v2/internal/application/chat"
	"youmind-backend-v2/internal/application/project"
	w6app "youmind-backend-v2/internal/application/w6"
	"youmind-backend-v2/internal/application/skill"
	"youmind-backend-v2/internal/application/user"
	"youmind-backend-v2/internal/config"
	"youmind-backend-v2/internal/infrastructure/ai"
	"youmind-backend-v2/internal/infrastructure/persistence"
)

// Wire 组装路由与依赖（可后续改为 wire/codegen）
func Wire(cfg *config.Config, db *persistence.DB) *gin.Engine {
	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowAllOrigins: true,
		AllowMethods:    []string{"*"},
		AllowHeaders:    []string{"*"},
	}))

	r.GET("/health", Health)

	api := r.Group("/api")
	userRepo := persistence.NewUserRepository(db)
	authSvc := auth.NewService(cfg, userRepo)
	authHandler := NewAuthHandler(authSvc, cfg)
	authHandler.RegisterRoutes(api.Group("/auth"))

	projectRepo := persistence.NewProjectRepository(db)
	sessionRepo := persistence.NewSessionRepository(db)
	messageRepo := persistence.NewMessageRepository(db)
	resourceRepo := persistence.NewResourceRepository(db)
	projectSvc := project.NewService(projectRepo, sessionRepo, messageRepo, resourceRepo)
	// W6 page maker (optional; only active when configured).
	w6Client := ai.NewW6Client(cfg.W6)
	w6WS := ai.NewW6WS(cfg.W6)
	pageMakerSvc := w6app.NewPageMakerService(w6Client, w6WS, resourceRepo)
	projectHandler := NewProjectHandler(projectSvc, pageMakerSvc)
	projectsGroup := api.Group("/projects")
	projectsGroup.Use(AuthMiddleware(authSvc))
	projectHandler.RegisterRoutes(projectsGroup)

	aiClient := ai.NewFromEnv()
	chatSvc := chat.NewService(projectRepo, sessionRepo, messageRepo, aiClient)
	chatHandler := NewChatHandler(chatSvc)
	chatGroup := api.Group("/chat")
	chatGroup.Use(AuthMiddleware(authSvc))
	chatHandler.RegisterRoutes(chatGroup)

	skillRepo := persistence.NewSkillRepository(db)
	skillSvc := skill.NewService(skillRepo)
	skillHandler := NewSkillHandler(skillSvc)
	skillsGroup := api.Group("/skills")
	skillsGroup.Use(AuthMiddleware(authSvc))
	skillHandler.RegisterRoutes(skillsGroup)

	modelsGroup := api.Group("/models")
	modelsHandler := NewModelsHandler()
	modelsHandler.RegisterRoutes(modelsGroup)

	userSvc := user.NewService(userRepo)
	userHandler := NewUserHandler(userSvc)
	userGroup := api.Group("/user")
	userGroup.Use(AuthMiddleware(authSvc))
	userHandler.RegisterRoutes(userGroup)

	// 静态文件服务 - 前端集成
	r.NoRoute(serveSPA)

	return r
}

// serveSPA 为 SPA 应用提供前端文件服务
func serveSPA(c *gin.Context) {
	path := filepath.Join("static", c.Request.URL.Path)
	if _, err := http.Dir(".").Open(path); err == nil {
		c.File(path)
		return
	}
	// 回退到 index.html，让前端路由处理
	c.File("static/index.html")
}
