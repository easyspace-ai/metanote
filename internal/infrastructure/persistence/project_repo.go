package persistence

import (
	"youmind-backend-v2/internal/domain/project"
)

// ProjectRepository 项目仓储 GORM 实现
type ProjectRepository struct {
	db *DB
}

func NewProjectRepository(db *DB) project.ProjectRepository {
	return &ProjectRepository{db: db}
}

func (r *ProjectRepository) Create(p *project.Project) error {
	m := toProjectModel(p)
	return r.db.Create(m).Error
}

func (r *ProjectRepository) GetByID(id string) (*project.Project, error) {
	var m ProjectModel
	if err := r.db.Where("id = ?", id).First(&m).Error; err != nil {
		return nil, err
	}
	return toProjectEntity(&m), nil
}

func (r *ProjectRepository) GetByIDAndUserID(id, userID string) (*project.Project, error) {
	var m ProjectModel
	if err := r.db.Where("id = ? AND user_id = ?", id, userID).First(&m).Error; err != nil {
		return nil, err
	}
	return toProjectEntity(&m), nil
}

func (r *ProjectRepository) ListByUserID(userID string, status *string, skip, limit int) ([]*project.Project, error) {
	var list []ProjectModel
	q := r.db.Where("user_id = ?", userID)
	if status != nil && *status != "" {
		q = q.Where("status = ?", *status)
	}
	if err := q.Order("updated_at DESC").Offset(skip).Limit(limit).Find(&list).Error; err != nil {
		return nil, err
	}
	out := make([]*project.Project, len(list))
	for i := range list {
		out[i] = toProjectEntity(&list[i])
	}
	return out, nil
}

func (r *ProjectRepository) Update(p *project.Project) error {
	m := toProjectModel(p)
	return r.db.Save(m).Error
}

func (r *ProjectRepository) Delete(id, userID string) error {
	res := r.db.Where("id = ? AND user_id = ?", id, userID).Delete(&ProjectModel{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gormErrNotFound
	}
	return nil
}

func toProjectModel(p *project.Project) *ProjectModel {
	return &ProjectModel{
		ID:          p.ID,
		UserID:      p.UserID,
		Name:        p.Name,
		Description: p.Description,
		CoverImage:  p.CoverImage,
		Status:      p.Status,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}

func toProjectEntity(m *ProjectModel) *project.Project {
	return &project.Project{
		ID:          m.ID,
		UserID:      m.UserID,
		Name:        m.Name,
		Description: m.Description,
		CoverImage:  m.CoverImage,
		Status:      m.Status,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}
