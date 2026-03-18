package project

import (
	"time"

	"github.com/google/uuid"
	"youmind-backend-v2/internal/domain/project"
)

// Service 项目应用服务
type Service struct {
	projectRepo  project.ProjectRepository
	sessionRepo  project.SessionRepository
	messageRepo  project.MessageRepository
	resourceRepo project.ResourceRepository
}

func NewService(
	projectRepo project.ProjectRepository,
	sessionRepo project.SessionRepository,
	messageRepo project.MessageRepository,
	resourceRepo project.ResourceRepository,
) *Service {
	return &Service{
		projectRepo:  projectRepo,
		sessionRepo:  sessionRepo,
		messageRepo:  messageRepo,
		resourceRepo: resourceRepo,
	}
}

// ListProjects 列出用户项目
func (s *Service) ListProjects(userID string, status *string, skip, limit int) ([]*project.Project, error) {
	return s.projectRepo.ListByUserID(userID, status, skip, limit)
}

// GetProject 获取单个项目（校验归属）
func (s *Service) GetProject(projectID, userID string) (*project.Project, error) {
	return s.projectRepo.GetByIDAndUserID(projectID, userID)
}

// CreateProject 创建项目
func (s *Service) CreateProject(userID, name string, description, coverImage *string) (*project.Project, error) {
	now := time.Now().UTC()
	p := &project.Project{
		ID:          uuid.NewString(),
		UserID:      userID,
		Name:        name,
		Description: description,
		CoverImage:  coverImage,
		Status:      "active",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := s.projectRepo.Create(p); err != nil {
		return nil, err
	}
	return p, nil
}

// UpdateProject 更新项目
func (s *Service) UpdateProject(projectID, userID string, name, description, coverImage, status *string) (*project.Project, error) {
	p, err := s.projectRepo.GetByIDAndUserID(projectID, userID)
	if err != nil {
		return nil, err
	}
	if name != nil {
		p.Name = *name
	}
	if description != nil {
		p.Description = description
	}
	if coverImage != nil {
		p.CoverImage = coverImage
	}
	if status != nil {
		p.Status = *status
	}
	p.UpdatedAt = time.Now().UTC()
	if err := s.projectRepo.Update(p); err != nil {
		return nil, err
	}
	return p, nil
}

// DeleteProject 删除项目
func (s *Service) DeleteProject(projectID, userID string) error {
	return s.projectRepo.Delete(projectID, userID)
}

// ListSessions 列出项目下的会话
func (s *Service) ListSessions(projectID string, skip, limit int) ([]*project.Session, error) {
	return s.sessionRepo.ListByProjectID(projectID, skip, limit)
}

// CreateSession 在项目下创建会话
func (s *Service) CreateSession(projectID, title string) (*project.Session, error) {
	now := time.Now().UTC()
	sess := &project.Session{
		ID:        uuid.NewString(),
		ProjectID: projectID,
		Title:     title,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.sessionRepo.Create(sess); err != nil {
		return nil, err
	}
	return sess, nil
}

// UpdateSession 更新会话标题
func (s *Service) UpdateSession(projectID, sessionID string, title string) (*project.Session, error) {
	sess, err := s.sessionRepo.GetByIDAndProjectID(sessionID, projectID)
	if err != nil {
		return nil, err
	}
	sess.Title = title
	sess.UpdatedAt = time.Now().UTC()
	if err := s.sessionRepo.Update(sess); err != nil {
		return nil, err
	}
	return sess, nil
}

// DeleteSession 删除会话（及其消息）
func (s *Service) DeleteSession(projectID, sessionID string) error {
	return s.sessionRepo.Delete(sessionID, projectID)
}

// GetSession 获取会话（校验归属项目）
func (s *Service) GetSession(projectID, sessionID string) (*project.Session, error) {
	return s.sessionRepo.GetByIDAndProjectID(sessionID, projectID)
}

// ListMessages 列出项目消息（兼容旧接口，按项目维度）
func (s *Service) ListMessages(projectID string, skip, limit int) ([]*project.Message, error) {
	return s.messageRepo.ListByProjectID(projectID, skip, limit)
}

// ListMessagesBySession 列出某会话的消息
func (s *Service) ListMessagesBySession(projectID, sessionID string, skip, limit int) ([]*project.Message, error) {
	if _, err := s.sessionRepo.GetByIDAndProjectID(sessionID, projectID); err != nil {
		return nil, err
	}
	return s.messageRepo.ListBySessionID(sessionID, skip, limit)
}

// CreateMessage 创建用户消息（需指定会话）
func (s *Service) CreateMessage(projectID, sessionID, content string, skillID *string, attachments map[string]interface{}) (*project.Message, error) {
	if _, err := s.sessionRepo.GetByIDAndProjectID(sessionID, projectID); err != nil {
		return nil, err
	}
	m := &project.Message{
		ID:         uuid.NewString(),
		ProjectID:  projectID,
		SessionID:  sessionID,
		Role:       "user",
		Content:    content,
		SkillID:    skillID,
		Attachments: attachments,
		CreatedAt:  time.Now().UTC(),
	}
	if err := s.messageRepo.Create(m); err != nil {
		return nil, err
	}
	return m, nil
}

// UpdateMessage 更新消息内容
func (s *Service) UpdateMessage(projectID, messageID, content string) (*project.Message, error) {
	return s.messageRepo.UpdateContent(projectID, messageID, content)
}

// DeleteMessage 删除消息
func (s *Service) DeleteMessage(projectID, messageID string) error {
	return s.messageRepo.Delete(projectID, messageID)
}

// ListResources 列出项目资源
func (s *Service) ListResources(projectID string, resourceType *string) ([]*project.Resource, error) {
	return s.resourceRepo.ListByProjectID(projectID, resourceType)
}

// CreateResource 创建资源
func (s *Service) CreateResource(projectID, resType, name string, content, url, size *string) (*project.Resource, error) {
	r := &project.Resource{
		ID:        uuid.NewString(),
		ProjectID: projectID,
		Type:      resType,
		Name:      name,
		Content:   content,
		URL:       url,
		Size:      size,
		CreatedAt: time.Now().UTC(),
	}
	if err := s.resourceRepo.Create(r); err != nil {
		return nil, err
	}
	return r, nil
}

// UpdateResource 更新资源
func (s *Service) UpdateResource(projectID, resourceID string, name, content, url *string) (*project.Resource, error) {
	r, err := s.resourceRepo.GetByID(projectID, resourceID)
	if err != nil {
		return nil, err
	}
	if name != nil {
		r.Name = *name
	}
	if content != nil {
		r.Content = content
	}
	if url != nil {
		r.URL = url
	}
	if err := s.resourceRepo.Update(r); err != nil {
		return nil, err
	}
	return r, nil
}

// DeleteResource 删除资源
func (s *Service) DeleteResource(projectID, resourceID string) error {
	return s.resourceRepo.Delete(projectID, resourceID)
}

// EnsureProjectBelongsToUser 校验项目归属，返回 nil 表示属于该用户
func (s *Service) EnsureProjectBelongsToUser(projectID, userID string) error {
	_, err := s.projectRepo.GetByIDAndUserID(projectID, userID)
	return err
}

// GetResource 获取单个资源
func (s *Service) GetResource(projectID, resourceID string) (*project.Resource, error) {
	return s.resourceRepo.GetByID(projectID, resourceID)
}
