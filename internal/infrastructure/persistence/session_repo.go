package persistence

import (
	"gorm.io/gorm"

	"youmind-backend-v2/internal/domain/project"
)

// SessionRepository 会话仓储 GORM 实现
type SessionRepository struct {
	db *DB
}

func NewSessionRepository(db *DB) project.SessionRepository {
	return &SessionRepository{db: db}
}

func (r *SessionRepository) Create(s *project.Session) error {
	m := toSessionModel(s)
	return r.db.Create(m).Error
}

func (r *SessionRepository) GetByID(id string) (*project.Session, error) {
	var m SessionModel
	if err := r.db.Where("id = ?", id).First(&m).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, gormErrNotFound
		}
		return nil, err
	}
	return toSessionEntity(&m), nil
}

func (r *SessionRepository) GetByIDAndProjectID(id, projectID string) (*project.Session, error) {
	var m SessionModel
	if err := r.db.Where("id = ? AND project_id = ?", id, projectID).First(&m).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, gormErrNotFound
		}
		return nil, err
	}
	return toSessionEntity(&m), nil
}

func (r *SessionRepository) ListByProjectID(projectID string, skip, limit int) ([]*project.Session, error) {
	var list []SessionModel
	if err := r.db.Where("project_id = ?", projectID).Order("updated_at DESC").Offset(skip).Limit(limit).Find(&list).Error; err != nil {
		return nil, err
	}
	out := make([]*project.Session, len(list))
	for i := range list {
		out[i] = toSessionEntity(&list[i])
	}
	return out, nil
}

func (r *SessionRepository) Update(s *project.Session) error {
	m := toSessionModel(s)
	return r.db.Save(m).Error
}

func (r *SessionRepository) Delete(id, projectID string) error {
	res := r.db.Where("id = ? AND project_id = ?", id, projectID).Delete(&SessionModel{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gormErrNotFound
	}
	return nil
}

func toSessionModel(s *project.Session) *SessionModel {
	return &SessionModel{
		ID:        s.ID,
		ProjectID: s.ProjectID,
		Title:     s.Title,
		CreatedAt: s.CreatedAt,
		UpdatedAt: s.UpdatedAt,
	}
}

func toSessionEntity(m *SessionModel) *project.Session {
	return &project.Session{
		ID:        m.ID,
		ProjectID: m.ProjectID,
		Title:     m.Title,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}
