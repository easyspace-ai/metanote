package persistence

import (
	"time"

	"github.com/google/uuid"
	"youmind-backend-v2/internal/domain/skill"
)

// SkillRepository 技能仓储 GORM 实现
type SkillRepository struct {
	db *DB
}

func NewSkillRepository(db *DB) skill.Repository {
	return &SkillRepository{db: db}
}

func (r *SkillRepository) Create(s *skill.Skill) error {
	m := toSkillModel(s)
	if m.ID == "" {
		m.ID = uuid.NewString()
	}
	now := time.Now().UTC()
	m.CreatedAt = now
	m.UpdatedAt = now
	return r.db.Create(m).Error
}

func (r *SkillRepository) GetByID(id string) (*skill.Skill, error) {
	var m SkillModel
	if err := r.db.Where("id = ?", id).First(&m).Error; err != nil {
		return nil, err
	}
	return toSkillEntity(&m), nil
}

func (r *SkillRepository) SetInstalled(id string, installed bool) error {
	res := r.db.Model(&SkillModel{}).Where("id = ?", id).Update("is_installed", installed)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gormErrNotFound
	}
	return nil
}

func (r *SkillRepository) ListAll() ([]*skill.Skill, error) {
	var list []SkillModel
	if err := r.db.Find(&list).Error; err != nil {
		return nil, err
	}
	out := make([]*skill.Skill, len(list))
	for i := range list {
		out[i] = toSkillEntity(&list[i])
	}
	return out, nil
}

func (r *SkillRepository) ListInstalled() ([]*skill.Skill, error) {
	var list []SkillModel
	if err := r.db.Where("is_installed = ?", true).Find(&list).Error; err != nil {
		return nil, err
	}
	out := make([]*skill.Skill, len(list))
	for i := range list {
		out[i] = toSkillEntity(&list[i])
	}
	return out, nil
}

func (r *SkillRepository) ListRecommended(limit int) ([]*skill.Skill, error) {
	if limit <= 0 {
		limit = 4
	}
	var list []SkillModel
	if err := r.db.Where("is_recommended = ?", true).Order("users_count DESC").Limit(limit).Find(&list).Error; err != nil {
		return nil, err
	}
	out := make([]*skill.Skill, len(list))
	for i := range list {
		out[i] = toSkillEntity(&list[i])
	}
	return out, nil
}

func toSkillModel(s *skill.Skill) *SkillModel {
	m := &SkillModel{
		ID:            s.ID,
		Name:          s.Name,
		Description:   s.Description,
		Icon:          s.Icon,
		Category:      s.Category,
		Author:        s.Author,
		UsersCount:    s.UsersCount,
		Rating:        s.Rating,
		Tags:          JSONSlice(s.Tags),
		SystemPrompt:  s.SystemPrompt,
		IsInstalled:   s.IsInstalled,
		IsPersonal:    s.IsPersonal,
		IsRecommended: s.IsRecommended,
	}
	if m.Category == "" {
		m.Category = "other"
	}
	return m
}

func toSkillEntity(m *SkillModel) *skill.Skill {
	return &skill.Skill{
		ID:            m.ID,
		Name:          m.Name,
		Description:   m.Description,
		Icon:          m.Icon,
		Category:      m.Category,
		Author:        m.Author,
		UsersCount:    m.UsersCount,
		Rating:        m.Rating,
		Tags:          []string(m.Tags),
		SystemPrompt:  m.SystemPrompt,
		IsInstalled:   m.IsInstalled,
		IsPersonal:    m.IsPersonal,
		IsRecommended: m.IsRecommended,
	}
}
