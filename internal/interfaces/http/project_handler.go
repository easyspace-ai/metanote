package http

import (
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"youmind-backend-v2/internal/application/project"
	w6app "youmind-backend-v2/internal/application/w6"
	projectdomain "youmind-backend-v2/internal/domain/project"
)

// ProjectHandler 项目/消息/资源/上传 HTTP 处理
type ProjectHandler struct {
	svc       *project.Service
	pageMaker *w6app.PageMakerService
}

func NewProjectHandler(svc *project.Service, pageMaker *w6app.PageMakerService) *ProjectHandler {
	return &ProjectHandler{svc: svc, pageMaker: pageMaker}
}

func (h *ProjectHandler) RegisterRoutes(r *gin.RouterGroup) {
	r.GET("", h.listProjects)
	r.POST("", h.createProject)
	r.GET("/:project_id", h.getProject)
	r.PATCH("/:project_id", h.updateProject)
	r.DELETE("/:project_id", h.deleteProject)
	// 会话：一个项目下多个会话
	r.GET("/:project_id/sessions", h.listSessions)
	r.POST("/:project_id/sessions", h.createSession)
	r.PATCH("/:project_id/sessions/:session_id", h.updateSession)
	r.DELETE("/:project_id/sessions/:session_id", h.deleteSession)
	r.GET("/:project_id/sessions/:session_id/messages", h.listMessagesBySession)
	// 消息（按项目维度保留兼容；按会话维度用上面）
	r.GET("/:project_id/messages", h.listMessages)
	r.POST("/:project_id/messages", h.createMessage)
	r.PATCH("/:project_id/messages/:message_id", h.updateMessage)
	r.DELETE("/:project_id/messages/:message_id", h.deleteMessage)
	r.GET("/:project_id/resources", h.listResources)
	r.POST("/:project_id/resources", h.createResource)
	r.PATCH("/:project_id/resources/:resource_id", h.updateResource)
	r.DELETE("/:project_id/resources/:resource_id", h.deleteResource)
	r.POST("/:project_id/upload", h.uploadFile)
	r.POST("/:project_id/page-from-outline", h.generatePageFromOutline)
}

func parseSkipLimit(c *gin.Context) (skip, limit int) {
	skip, limit = 0, 20
	if v := c.Query("skip"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			skip = n
		}
	}
	if v := c.Query("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			limit = n
		}
	}
	return skip, limit
}

func (h *ProjectHandler) listProjects(c *gin.Context) {
	u, ok := GetCurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"detail": "Not authenticated"})
		return
	}
	status := c.Query("status")
	var statusPtr *string
	if status != "" {
		statusPtr = &status
	}
	skip, limit := parseSkipLimit(c)
	list, err := h.svc.ListProjects(u.ID, statusPtr, skip, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "failed to list projects"})
		return
	}
	c.JSON(http.StatusOK, toProjectListResponse(list))
}

func (h *ProjectHandler) getProject(c *gin.Context) {
	u, ok := GetCurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"detail": "Not authenticated"})
		return
	}
	p, err := h.svc.GetProject(c.Param("project_id"), u.ID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"detail": "Project not found"})
		return
	}
	c.JSON(http.StatusOK, toProjectResponse(p))
}

func (h *ProjectHandler) createProject(c *gin.Context) {
	u, ok := GetCurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"detail": "Not authenticated"})
		return
	}
	var req struct {
		Name        string  `json:"name" binding:"required"`
		Description *string `json:"description"`
		CoverImage  *string `json:"cover_image"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "invalid request body"})
		return
	}
	p, err := h.svc.CreateProject(u.ID, req.Name, req.Description, req.CoverImage)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "failed to create project"})
		return
	}
	c.JSON(http.StatusOK, toProjectResponse(p))
}

func (h *ProjectHandler) updateProject(c *gin.Context) {
	u, ok := GetCurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"detail": "Not authenticated"})
		return
	}
	var req struct {
		Name        *string `json:"name"`
		Description *string `json:"description"`
		CoverImage  *string `json:"cover_image"`
		Status      *string `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "invalid request body"})
		return
	}
	p, err := h.svc.UpdateProject(c.Param("project_id"), u.ID, req.Name, req.Description, req.CoverImage, req.Status)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"detail": "Project not found"})
		return
	}
	c.JSON(http.StatusOK, toProjectResponse(p))
}

func (h *ProjectHandler) deleteProject(c *gin.Context) {
	u, ok := GetCurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"detail": "Not authenticated"})
		return
	}
	if err := h.svc.DeleteProject(c.Param("project_id"), u.ID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"detail": "Project not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Project deleted"})
}

func toProjectResponse(p *projectdomain.Project) gin.H {
	return gin.H{
		"id":          p.ID,
		"name":        p.Name,
		"description": p.Description,
		"cover_image": p.CoverImage,
		"status":      p.Status,
		"created_at":  p.CreatedAt,
		"updated_at":  p.UpdatedAt,
	}
}

func toProjectListResponse(list []*projectdomain.Project) []gin.H {
	out := make([]gin.H, len(list))
	for i, p := range list {
		out[i] = toProjectResponse(p)
	}
	return out
}

func (h *ProjectHandler) listSessions(c *gin.Context) {
	u, ok := GetCurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"detail": "Not authenticated"})
		return
	}
	projectID := c.Param("project_id")
	if err := h.svc.EnsureProjectBelongsToUser(projectID, u.ID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"detail": "Project not found"})
		return
	}
	skip, limit := parseSkipLimit(c)
	if limit > 50 {
		limit = 50
	}
	list, err := h.svc.ListSessions(projectID, skip, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "failed to list sessions"})
		return
	}
	c.JSON(http.StatusOK, toSessionListResponse(list))
}

func (h *ProjectHandler) createSession(c *gin.Context) {
	u, ok := GetCurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"detail": "Not authenticated"})
		return
	}
	projectID := c.Param("project_id")
	if err := h.svc.EnsureProjectBelongsToUser(projectID, u.ID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"detail": "Project not found"})
		return
	}
	var req struct {
		Title string `json:"title"`
	}
	_ = c.ShouldBindJSON(&req)
	if req.Title == "" {
		req.Title = "新对话"
	}
	sess, err := h.svc.CreateSession(projectID, req.Title)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "failed to create session"})
		return
	}
	c.JSON(http.StatusOK, toSessionResponse(sess))
}

func (h *ProjectHandler) updateSession(c *gin.Context) {
	u, ok := GetCurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"detail": "Not authenticated"})
		return
	}
	projectID := c.Param("project_id")
	sessionID := c.Param("session_id")
	if err := h.svc.EnsureProjectBelongsToUser(projectID, u.ID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"detail": "Project not found"})
		return
	}
	var req struct {
		Title string `json:"title" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "invalid request body"})
		return
	}
	sess, err := h.svc.UpdateSession(projectID, sessionID, req.Title)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"detail": "Session not found"})
		return
	}
	c.JSON(http.StatusOK, toSessionResponse(sess))
}

func (h *ProjectHandler) deleteSession(c *gin.Context) {
	u, ok := GetCurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"detail": "Not authenticated"})
		return
	}
	projectID := c.Param("project_id")
	sessionID := c.Param("session_id")
	if err := h.svc.EnsureProjectBelongsToUser(projectID, u.ID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"detail": "Project not found"})
		return
	}
	if err := h.svc.DeleteSession(projectID, sessionID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"detail": "Session not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Session deleted"})
}

func (h *ProjectHandler) listMessagesBySession(c *gin.Context) {
	u, ok := GetCurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"detail": "Not authenticated"})
		return
	}
	projectID := c.Param("project_id")
	sessionID := c.Param("session_id")
	if err := h.svc.EnsureProjectBelongsToUser(projectID, u.ID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"detail": "Project not found"})
		return
	}
	skip, limit := parseSkipLimit(c)
	if limit > 50 {
		limit = 50
	}
	list, err := h.svc.ListMessagesBySession(projectID, sessionID, skip, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "failed to list messages"})
		return
	}
	c.JSON(http.StatusOK, toMessageListResponse(list))
}

func toSessionResponse(s *projectdomain.Session) gin.H {
	return gin.H{
		"id":         s.ID,
		"project_id": s.ProjectID,
		"title":      s.Title,
		"created_at": s.CreatedAt,
		"updated_at": s.UpdatedAt,
	}
}

func toSessionListResponse(list []*projectdomain.Session) []gin.H {
	out := make([]gin.H, len(list))
	for i, s := range list {
		out[i] = toSessionResponse(s)
	}
	return out
}

func (h *ProjectHandler) listMessages(c *gin.Context) {
	projectID := c.Param("project_id")
	skip, limit := parseSkipLimit(c)
	if limit > 50 {
		limit = 50
	}
	list, err := h.svc.ListMessages(projectID, skip, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "failed to list messages"})
		return
	}
	c.JSON(http.StatusOK, toMessageListResponse(list))
}

func (h *ProjectHandler) createMessage(c *gin.Context) {
	u, ok := GetCurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"detail": "Not authenticated"})
		return
	}
	projectID := c.Param("project_id")
	if err := h.svc.EnsureProjectBelongsToUser(projectID, u.ID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"detail": "Project not found"})
		return
	}
	var req struct {
		SessionID   string                 `json:"session_id" binding:"required"`
		Content      string                 `json:"content" binding:"required"`
		SkillID      *string                `json:"skill_id"`
		Attachments  map[string]interface{} `json:"attachments"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "invalid request body"})
		return
	}
	m, err := h.svc.CreateMessage(projectID, req.SessionID, req.Content, req.SkillID, req.Attachments)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "failed to create message"})
		return
	}
	c.JSON(http.StatusOK, toMessageResponse(m))
}

func (h *ProjectHandler) updateMessage(c *gin.Context) {
	var req struct {
		Content string `json:"content" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "invalid request body"})
		return
	}
	m, err := h.svc.UpdateMessage(c.Param("project_id"), c.Param("message_id"), req.Content)
	if err != nil || m == nil {
		c.JSON(http.StatusNotFound, gin.H{"detail": "Message not found"})
		return
	}
	c.JSON(http.StatusOK, toMessageResponse(m))
}

func (h *ProjectHandler) deleteMessage(c *gin.Context) {
	if err := h.svc.DeleteMessage(c.Param("project_id"), c.Param("message_id")); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"detail": "Message not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Message deleted"})
}

func toMessageResponse(m *projectdomain.Message) gin.H {
	return gin.H{
		"id":         m.ID,
		"project_id": m.ProjectID,
		"session_id": m.SessionID,
		"role":       m.Role,
		"content":    m.Content,
		"skill_id":   m.SkillID,
		"attachments": m.Attachments,
		"created_at": m.CreatedAt,
	}
}

func toMessageListResponse(list []*projectdomain.Message) []gin.H {
	out := make([]gin.H, len(list))
	for i, m := range list {
		out[i] = toMessageResponse(m)
	}
	return out
}

func (h *ProjectHandler) listResources(c *gin.Context) {
	projectID := c.Param("project_id")
	typ := c.Query("type")
	var typPtr *string
	if typ != "" {
		typPtr = &typ
	}
	list, err := h.svc.ListResources(projectID, typPtr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "failed to list resources"})
		return
	}
	c.JSON(http.StatusOK, toResourceListResponse(list))
}

func (h *ProjectHandler) createResource(c *gin.Context) {
	u, ok := GetCurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"detail": "Not authenticated"})
		return
	}
	projectID := c.Param("project_id")
	if err := h.svc.EnsureProjectBelongsToUser(projectID, u.ID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"detail": "Project not found"})
		return
	}
	var req struct {
		Type    string  `json:"type" binding:"required"`
		Name    string  `json:"name" binding:"required"`
		Content *string `json:"content"`
		URL     *string `json:"url"`
		Size    *string `json:"size"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "invalid request body"})
		return
	}
	r, err := h.svc.CreateResource(projectID, req.Type, req.Name, req.Content, req.URL, req.Size)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "failed to create resource"})
		return
	}
	c.JSON(http.StatusOK, toResourceResponse(r))
}

func (h *ProjectHandler) updateResource(c *gin.Context) {
	var req struct {
		Name    *string `json:"name"`
		Content *string `json:"content"`
		URL     *string `json:"url"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "invalid request body"})
		return
	}
	r, err := h.svc.UpdateResource(c.Param("project_id"), c.Param("resource_id"), req.Name, req.Content, req.URL)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"detail": "Resource not found"})
		return
	}
	c.JSON(http.StatusOK, toResourceResponse(r))
}

func (h *ProjectHandler) deleteResource(c *gin.Context) {
	if err := h.svc.DeleteResource(c.Param("project_id"), c.Param("resource_id")); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"detail": "Resource not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Resource deleted"})
}

func toResourceResponse(r *projectdomain.Resource) gin.H {
	return gin.H{
		"id":         r.ID,
		"project_id": r.ProjectID,
		"type":       r.Type,
		"name":       r.Name,
		"content":    r.Content,
		"url":        r.URL,
		"size":       r.Size,
		"created_at": r.CreatedAt,
	}
}

func toResourceListResponse(list []*projectdomain.Resource) []gin.H {
	out := make([]gin.H, len(list))
	for i, r := range list {
		out[i] = toResourceResponse(r)
	}
	return out
}

func (h *ProjectHandler) uploadFile(c *gin.Context) {
	u, ok := GetCurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"detail": "Not authenticated"})
		return
	}
	projectID := c.Param("project_id")
	if err := h.svc.EnsureProjectBelongsToUser(projectID, u.ID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"detail": "Project not found"})
		return
	}
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "file is required"})
		return
	}
	// 简单实现：保存到本地并创建 resource（与旧版一致）
	// 可后续抽成 application 的 UploadFile use case
	allowedExt := map[string]bool{
		".pdf": true, ".doc": true, ".docx": true, ".txt": true, ".md": true,
		".jpg": true, ".jpeg": true, ".png": true,
	}
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !allowedExt[ext] {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "unsupported file type"})
		return
	}
	uploadDir := "uploads"
	if err := os.MkdirAll(uploadDir, 0o755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "failed to create upload directory"})
		return
	}
	filename := uuid.NewString() + "_" + file.Filename
	targetPath := filepath.Join(uploadDir, filename)
	if err := c.SaveUploadedFile(file, targetPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "failed to save file"})
		return
	}
	sizeStr := strconv.FormatInt(file.Size, 10)
	r, err := h.svc.CreateResource(projectID, "document", file.Filename, nil, &targetPath, &sizeStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "failed to create resource for upload"})
		return
	}
	c.JSON(http.StatusOK, toResourceResponse(r))
}

// generatePageFromOutline 使用 W6 pagemaker 根据项目大纲生成 HTML 页面资源。
// 请求体可以直接携带 outline 文本，或提供 outline 资源 ID 从现有资源中读取。
type generatePageRequest struct {
	Title             string  `json:"title" binding:"required"`
	KnowledgePoints   string  `json:"knowledge_points"`
	Outline           *string `json:"outline,omitempty"`
	OutlineResourceID *string `json:"outline_resource_id,omitempty"`
}

func (h *ProjectHandler) generatePageFromOutline(c *gin.Context) {
	if h.pageMaker == nil {
		c.JSON(http.StatusNotImplemented, gin.H{"detail": "W6 pagemaker is not configured in backend"})
		return
	}

	u, ok := GetCurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"detail": "Not authenticated"})
		return
	}
	projectID := c.Param("project_id")
	if err := h.svc.EnsureProjectBelongsToUser(projectID, u.ID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"detail": "Project not found"})
		return
	}

	var req generatePageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "invalid request body"})
		return
	}

	var outlineText string
	if req.Outline != nil && *req.Outline != "" {
		outlineText = *req.Outline
	} else if req.OutlineResourceID != nil && *req.OutlineResourceID != "" {
		res, err := h.svc.GetResource(projectID, *req.OutlineResourceID)
		if err != nil || res == nil || res.Content == nil {
			c.JSON(http.StatusNotFound, gin.H{"detail": "Outline resource not found"})
			return
		}
		outlineText = *res.Content
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "outline or outline_resource_id is required"})
		return
	}

	ctx := c.Request.Context()

	// Stream progress via SSE when client sends Accept: text/event-stream or ?stream=1
	streamReq := c.Query("stream") == "1" || strings.Contains(c.GetHeader("Accept"), "text/event-stream")
	if streamReq {
		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")
		c.Header("X-Accel-Buffering", "no")
		flusher, ok := c.Writer.(http.Flusher)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"detail": "streaming not supported"})
			return
		}
		stepMessages := map[string]string{
			"created_chat":      "已创建对话",
			"calling_pagemaker": "正在调用 PageMaker 代理...",
			"waiting_artefact":  "正在生成网页，请稍候...",
			"got_artefact":      "已生成，正在拉取结果...",
			"saving":            "正在保存到项目...",
			"done":              "完成",
		}
		sendProgress := func(step string) {
			msg := stepMessages[step]
			if msg == "" {
				msg = step
			}
			c.SSEvent("progress", gin.H{"step": step, "message": msg})
			flusher.Flush()
		}
		res, err := h.pageMaker.GeneratePageFromOutline(ctx, projectID, req.Title, req.KnowledgePoints, outlineText, sendProgress)
		if err != nil {
			c.SSEvent("error", gin.H{"detail": err.Error()})
			flusher.Flush()
			return
		}
		c.SSEvent("result", gin.H{
			"id":         res.ID,
			"project_id": res.ProjectID,
			"type":       res.Type,
			"name":       res.Name,
			"content":    res.Content,
			"created_at": res.CreatedAt,
		})
		flusher.Flush()
		return
	}

	res, err := h.pageMaker.GeneratePageFromOutline(ctx, projectID, req.Title, req.KnowledgePoints, outlineText)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "failed to generate page: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":         res.ID,
		"project_id": res.ProjectID,
		"type":       res.Type,
		"name":       res.Name,
		"content":    res.Content,
		"created_at": res.CreatedAt,
	})
}
