package skill

import (
	"github.com/google/uuid"
	"youmind-backend-v2/internal/domain/skill"
)

// Service 技能应用服务
type Service struct {
	repo skill.Repository
}

func NewService(repo skill.Repository) *Service {
	return &Service{repo: repo}
}

// CreateInput 创建技能入参（与前端 / Python SkillCreate 对齐）
type CreateInput struct {
	Name         string
	Description  *string
	Icon         *string
	Category     string
	SystemPrompt *string
}

// Create 创建新技能（个人技能 is_personal=true）
func (s *Service) Create(in CreateInput) (*skill.Skill, error) {
	if in.Category == "" {
		in.Category = "other"
	}
	desc := in.Description
	prompt := in.SystemPrompt
	if prompt == nil && desc != nil {
		prompt = desc
	}
	sk := &skill.Skill{
		ID:           uuid.NewString(),
		Name:         in.Name,
		Description:  desc,
		Icon:         in.Icon,
		Category:     in.Category,
		SystemPrompt: prompt,
		IsPersonal:   true,
	}
	if err := s.repo.Create(sk); err != nil {
		return nil, err
	}
	return sk, nil
}

// GetByID 获取单个技能
func (s *Service) GetByID(id string) (*skill.Skill, error) {
	return s.repo.GetByID(id)
}

// Install 将技能设为已安装（持久化到库）
func (s *Service) Install(id string) error {
	return s.repo.SetInstalled(id, true)
}

// Uninstall 将技能设为未安装
func (s *Service) Uninstall(id string) error {
	return s.repo.SetInstalled(id, false)
}

// ListAll 列出所有技能
func (s *Service) ListAll() ([]*skill.Skill, error) {
	return s.repo.ListAll()
}

// ListInstalled 列出已安装技能
func (s *Service) ListInstalled() ([]*skill.Skill, error) {
	return s.repo.ListInstalled()
}

// ListRecommended 列出推荐技能
func (s *Service) ListRecommended(limit int) ([]*skill.Skill, error) {
	return s.repo.ListRecommended(limit)
}
