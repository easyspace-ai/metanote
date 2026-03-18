package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"youmind-backend-v2/internal/application/skill"
	skilldomain "youmind-backend-v2/internal/domain/skill"
)

// SkillHandler 技能 HTTP 处理
type SkillHandler struct {
	svc *skill.Service
}

func NewSkillHandler(svc *skill.Service) *SkillHandler {
	return &SkillHandler{svc: svc}
}

func (h *SkillHandler) RegisterRoutes(r *gin.RouterGroup) {
	r.GET("", h.listSkills)
	r.POST("", h.createSkill)
	r.GET("/installed", h.listInstalled)
	r.GET("/recommended", h.listRecommended)
	r.GET("/:id", h.getSkill)
	r.POST("/:id/install", h.installSkill)
	r.POST("/:id/uninstall", h.uninstallSkill)
}

func (h *SkillHandler) listSkills(c *gin.Context) {
	list, err := h.svc.ListAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "failed to list skills"})
		return
	}
	c.JSON(http.StatusOK, toSkillListResponse(list))
}

type createSkillRequest struct {
	Name         string  `json:"name" binding:"required"`
	Description  *string `json:"description"`
	Icon         *string `json:"icon"`
	Category     string  `json:"category"`
	SystemPrompt *string `json:"system_prompt"`
}

func (h *SkillHandler) createSkill(c *gin.Context) {
	var req createSkillRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "invalid request body"})
		return
	}
	in := skill.CreateInput{
		Name:         req.Name,
		Description:  req.Description,
		Icon:         req.Icon,
		Category:     req.Category,
		SystemPrompt: req.SystemPrompt,
	}
	sk, err := h.svc.Create(in)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "failed to create skill"})
		return
	}
	c.JSON(http.StatusCreated, toSkillResponse(sk))
}

func (h *SkillHandler) getSkill(c *gin.Context) {
	id := c.Param("id")
	sk, err := h.svc.GetByID(id)
	if err != nil || sk == nil {
		c.JSON(http.StatusNotFound, gin.H{"detail": "Skill not found"})
		return
	}
	c.JSON(http.StatusOK, toSkillResponse(sk))
}

func (h *SkillHandler) listInstalled(c *gin.Context) {
	list, err := h.svc.ListInstalled()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "failed to list installed skills"})
		return
	}
	c.JSON(http.StatusOK, toSkillListResponse(list))
}

func (h *SkillHandler) listRecommended(c *gin.Context) {
	list, err := h.svc.ListRecommended(4)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "failed to list recommended skills"})
		return
	}
	c.JSON(http.StatusOK, toSkillListResponse(list))
}

func (h *SkillHandler) installSkill(c *gin.Context) {
	id := c.Param("id")
	if err := h.svc.Install(id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"detail": "Skill not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}

func (h *SkillHandler) uninstallSkill(c *gin.Context) {
	id := c.Param("id")
	if err := h.svc.Uninstall(id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"detail": "Skill not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}

func toSkillResponse(s *skilldomain.Skill) gin.H {
	return gin.H{
		"id":             s.ID,
		"name":           s.Name,
		"description":    s.Description,
		"icon":           s.Icon,
		"category":       s.Category,
		"author":         s.Author,
		"users_count":    s.UsersCount,
		"rating":         s.Rating,
		"tags":           s.Tags,
		"system_prompt":  s.SystemPrompt,
		"is_installed":   s.IsInstalled,
		"is_personal":    s.IsPersonal,
		"is_recommended": s.IsRecommended,
	}
}

func toSkillListResponse(list []*skilldomain.Skill) []gin.H {
	out := make([]gin.H, len(list))
	for i, s := range list {
		out[i] = toSkillResponse(s)
	}
	return out
}
