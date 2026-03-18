package http

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"youmind-backend-v2/internal/application/chat"
)

// ChatHandler 对话 HTTP 处理
type ChatHandler struct {
	svc *chat.Service
}

func NewChatHandler(svc *chat.Service) *ChatHandler {
	return &ChatHandler{svc: svc}
}

func (h *ChatHandler) RegisterRoutes(r *gin.RouterGroup) {
	r.POST("", h.chat)
	r.POST("/stream", h.chatStream)
}

type chatRequest struct {
	Message     string                 `json:"message" binding:"required"`
	ProjectID   *string                `json:"project_id"`   // 必填，会话归属项目
	SessionID   *string                `json:"session_id"`  // 可选，不传则新建会话
	SkillID     *string                `json:"skill_id"`
	Attachments map[string]interface{} `json:"attachments"`
	Model       *string                `json:"model"`
	Mode        *string                `json:"mode"`
}

func (h *ChatHandler) chat(c *gin.Context) {
	u, ok := GetCurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"detail": "Not authenticated"})
		return
	}
	var req chatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "invalid request body"})
		return
	}
	result, err := h.svc.Chat(c.Request.Context(), u.ID, chat.ChatInput{
		Message:     req.Message,
		ProjectID:   req.ProjectID,
		SessionID:   req.SessionID,
		SkillID:     req.SkillID,
		Attachments: req.Attachments,
		Model:       req.Model,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "ai error: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"id":         result.ID,
		"project_id": result.ProjectID,
		"session_id": result.SessionID,
		"role":       result.Role,
		"content":    result.Content,
		"skill_id":   result.SkillID,
		"created_at": result.CreatedAt,
	})
}

// chatStream 流式对话：先鉴权、解析/创建会话、保存用户消息并自动更新会话标题，再流式返回，最后持久化助手消息。
func (h *ChatHandler) chatStream(c *gin.Context) {
	u, ok := GetCurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"detail": "Not authenticated"})
		return
	}
	var req chatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "invalid request body"})
		return
	}
	if req.ProjectID == nil || *req.ProjectID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "project_id is required"})
		return
	}

	sessionID, err := h.svc.PrepareSessionAndSaveUserMessage(c.Request.Context(), u.ID, chat.ChatInput{
		Message:     req.Message,
		ProjectID:   req.ProjectID,
		SessionID:   req.SessionID,
		SkillID:     req.SkillID,
		Attachments: req.Attachments,
		Model:       req.Model,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "prepare session: " + err.Error()})
		return
	}

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "streaming not supported"})
		return
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	// 读取 OpenAI 兼容配置
	baseURL := os.Getenv("OPENAI_COMPAT_BASE_URL")
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}
	apiKey := os.Getenv("OPENAI_COMPAT_API_KEY")
	model := os.Getenv("OPENAI_COMPAT_MODEL")
	if model == "" {
		model = "gpt-4.1-mini"
	}

	// 构造上游流式请求
	type chatMessage struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}
	type streamRequest struct {
		Model    string        `json:"model"`
		Messages []chatMessage `json:"messages"`
		Stream   bool          `json:"stream"`
	}

	upReqBody := streamRequest{
		Model: model,
		Messages: []chatMessage{
			{Role: "user", Content: req.Message},
		},
		Stream: true,
	}
	bodyBytes, err := json.Marshal(upReqBody)
	if err != nil {
		fmt.Fprintf(c.Writer, "data: %s\n\n", `{"type":"error","value":"marshal upstream request failed"}`)
		flusher.Flush()
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Minute)
	defer cancel()

	upReq, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(baseURL, "/")+"/chat/completions", strings.NewReader(string(bodyBytes)))
	if err != nil {
		msg := fmt.Sprintf(`{"type":"error","value":"build upstream request failed: %v"}`, err)
		fmt.Fprintf(c.Writer, "data: %s\n\n", msg)
		flusher.Flush()
		return
	}
	upReq.Header.Set("Content-Type", "application/json")
	if apiKey != "" {
		upReq.Header.Set("Authorization", "Bearer "+apiKey)
	}

	httpClient := &http.Client{Timeout: 0} // 由上面的 context 控制超时
	resp, err := httpClient.Do(upReq)
	if err != nil {
		msg := fmt.Sprintf(`{"type":"error","value":"call upstream chat api failed: %v"}`, err)
		fmt.Fprintf(c.Writer, "data: %s\n\n", msg)
		flusher.Flush()
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		msg := fmt.Sprintf(`{"type":"error","value":"upstream chat api status %d"}`, resp.StatusCode)
		fmt.Fprintf(c.Writer, "data: %s\n\n", msg)
		flusher.Flush()
		return
	}

	// OpenAI 兼容流式增量响应结构
	type delta struct {
		Content string `json:"content"`
	}
	type choice struct {
		Delta delta `json:"delta"`
	}
	type streamResp struct {
		Choices []choice `json:"choices"`
	}

	var fullContent strings.Builder
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if !strings.HasPrefix(line, "data:") {
			continue
		}
		data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if data == "" || data == "[DONE]" {
			if data == "[DONE]" {
				break
			}
			continue
		}

		var up streamResp
		if err := json.Unmarshal([]byte(data), &up); err != nil {
			payload := fmt.Sprintf(`{"type":"content","value":%q}`, data)
			fmt.Fprintf(c.Writer, "data: %s\n\n", payload)
			fullContent.WriteString(data)
			flusher.Flush()
			continue
		}
		for _, ch := range up.Choices {
			if ch.Delta.Content == "" {
				continue
			}
			fullContent.WriteString(ch.Delta.Content)
			payload := fmt.Sprintf(`{"type":"content","value":%q}`, ch.Delta.Content)
			fmt.Fprintf(c.Writer, "data: %s\n\n", payload)
			flusher.Flush()
		}
	}

	// 流式结束后持久化助手消息
	if fullContent.Len() > 0 {
		_ = h.svc.SaveAssistantMessage(*req.ProjectID, sessionID, fullContent.String(), req.SkillID)
	}

	// 可选：通知前端会话已更新（便于刷新会话列表标题）
	fmt.Fprintf(c.Writer, "data: %s\n\n", `{"type":"session_id","value":`+fmt.Sprintf("%q", sessionID)+`}`)
	fmt.Fprintf(c.Writer, "data: %s\n\n", `{"type":"status_clear"}`)
	flusher.Flush()
}
