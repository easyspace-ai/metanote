package chat

import (
	"context"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"youmind-backend-v2/internal/domain/project"
)

// AIClient 抽象 AI 对话（由 infrastructure 实现）
type AIClient interface {
	Chat(ctx context.Context, userMessage string, modelOverride *string) (string, error)
}

// Service 对话应用服务
type Service struct {
	projectRepo project.ProjectRepository
	sessionRepo project.SessionRepository
	messageRepo project.MessageRepository
	aiClient    AIClient
}

func NewService(
	projectRepo project.ProjectRepository,
	sessionRepo project.SessionRepository,
	messageRepo project.MessageRepository,
	aiClient AIClient,
) *Service {
	return &Service{
		projectRepo: projectRepo,
		sessionRepo: sessionRepo,
		messageRepo: messageRepo,
		aiClient:    aiClient,
	}
}

// ChatInput 对话入参
type ChatInput struct {
	Message     string
	ProjectID   *string // 必填，会话归属项目
	SessionID   *string // 可选，不传则新建会话
	SkillID     *string
	Attachments map[string]interface{}
	Model       *string
}

// ChatResult 对话结果（返回助手消息）
type ChatResult struct {
	ID        string
	ProjectID string
	SessionID string
	Role      string
	Content   string
	SkillID   *string
	CreatedAt time.Time
}

// Chat 必须传入 project_id；可选 session_id。若无 session 则新建会话。消息归属到会话。
func (s *Service) Chat(ctx context.Context, userID string, in ChatInput) (*ChatResult, error) {
	if in.ProjectID == nil || *in.ProjectID == "" {
		return nil, fmt.Errorf("project_id is required")
	}
	projectID := *in.ProjectID
	if _, err := s.projectRepo.GetByIDAndUserID(projectID, userID); err != nil {
		return nil, err
	}

	sessionID := ""
	if in.SessionID != nil && *in.SessionID != "" {
		sess, err := s.sessionRepo.GetByIDAndProjectID(*in.SessionID, projectID)
		if err != nil {
			return nil, err
		}
		sessionID = *in.SessionID
		if sess.Title == "新对话" && strings.TrimSpace(in.Message) != "" {
			sess.Title = truncateTitle(in.Message, maxTitleRunes)
			sess.UpdatedAt = time.Now().UTC()
			_ = s.sessionRepo.Update(sess)
		}
	} else {
		now := time.Now().UTC()
		sess := &project.Session{
			ID:        uuid.NewString(),
			ProjectID: projectID,
			Title:     truncateTitle(in.Message, maxTitleRunes),
			CreatedAt: now,
			UpdatedAt: now,
		}
		if err := s.sessionRepo.Create(sess); err != nil {
			return nil, err
		}
		sessionID = sess.ID
	}

	userMsg := &project.Message{
		ID:         uuid.NewString(),
		ProjectID:  projectID,
		SessionID:  sessionID,
		Role:       "user",
		Content:    in.Message,
		SkillID:    in.SkillID,
		Attachments: in.Attachments,
		CreatedAt:  time.Now().UTC(),
	}
	if err := s.messageRepo.Create(userMsg); err != nil {
		return nil, err
	}

	assistantContent, err := s.aiClient.Chat(ctx, in.Message, in.Model)
	if err != nil {
		return nil, err
	}

	assistantMsg := &project.Message{
		ID:        uuid.NewString(),
		ProjectID: projectID,
		SessionID: sessionID,
		Role:      "assistant",
		Content:   assistantContent,
		SkillID:   in.SkillID,
		CreatedAt: time.Now().UTC(),
	}
	if err := s.messageRepo.Create(assistantMsg); err != nil {
		return nil, err
	}

	return &ChatResult{
		ID:        assistantMsg.ID,
		ProjectID: projectID,
		SessionID: sessionID,
		Role:      "assistant",
		Content:   assistantContent,
		SkillID:   in.SkillID,
		CreatedAt: assistantMsg.CreatedAt,
	}, nil
}

const maxTitleRunes = 28

// truncateTitle 截取为会话标题，最多 maxRunes 个字符
func truncateTitle(s string, maxRunes int) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return "新对话"
	}
	if utf8.RuneCountInString(s) <= maxRunes {
		return s
	}
	runes := []rune(s)
	return string(runes[:maxRunes]) + "…"
}

// PrepareSessionAndSaveUserMessage 解析/创建会话、保存用户消息，并视情况用首条消息更新会话标题。用于流式接口。
func (s *Service) PrepareSessionAndSaveUserMessage(ctx context.Context, userID string, in ChatInput) (sessionID string, err error) {
	if in.ProjectID == nil || *in.ProjectID == "" {
		return "", fmt.Errorf("project_id is required")
	}
	projectID := *in.ProjectID
	if _, err := s.projectRepo.GetByIDAndUserID(projectID, userID); err != nil {
		return "", err
	}

	titleFromMessage := truncateTitle(in.Message, maxTitleRunes)

	if in.SessionID != nil && *in.SessionID != "" {
		sess, err := s.sessionRepo.GetByIDAndProjectID(*in.SessionID, projectID)
		if err != nil {
			return "", err
		}
		sessionID = *in.SessionID
		// 若当前标题仍是「新对话」，用首条消息更新
		if sess.Title == "新对话" && strings.TrimSpace(in.Message) != "" {
			sess.Title = titleFromMessage
			sess.UpdatedAt = time.Now().UTC()
			_ = s.sessionRepo.Update(sess)
		}
	} else {
		now := time.Now().UTC()
		sess := &project.Session{
			ID:        uuid.NewString(),
			ProjectID: projectID,
			Title:     titleFromMessage,
			CreatedAt: now,
			UpdatedAt: now,
		}
		if err := s.sessionRepo.Create(sess); err != nil {
			return "", err
		}
		sessionID = sess.ID
	}

	userMsg := &project.Message{
		ID:         uuid.NewString(),
		ProjectID:  projectID,
		SessionID:  sessionID,
		Role:       "user",
		Content:    in.Message,
		SkillID:    in.SkillID,
		Attachments: in.Attachments,
		CreatedAt:  time.Now().UTC(),
	}
	if err := s.messageRepo.Create(userMsg); err != nil {
		return "", err
	}
	return sessionID, nil
}

// SaveAssistantMessage 流式结束后保存助手消息
func (s *Service) SaveAssistantMessage(projectID, sessionID, content string, skillID *string) error {
	_, err := s.sessionRepo.GetByIDAndProjectID(sessionID, projectID)
	if err != nil {
		return err
	}
	m := &project.Message{
		ID:        uuid.NewString(),
		ProjectID: projectID,
		SessionID: sessionID,
		Role:      "assistant",
		Content:   content,
		SkillID:   skillID,
		CreatedAt: time.Now().UTC(),
	}
	return s.messageRepo.Create(m)
}
