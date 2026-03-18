package persistence

import (
	"youmind-backend-v2/internal/domain/project"
)

// ResourceRepository 资源仓储 GORM 实现
type ResourceRepository struct {
	db *DB
}

func NewResourceRepository(db *DB) project.ResourceRepository {
	return &ResourceRepository{db: db}
}

func (r *ResourceRepository) Create(res *project.Resource) error {
	m := toResourceModel(res)
	return r.db.Create(m).Error
}

func (r *ResourceRepository) GetByID(projectID, resourceID string) (*project.Resource, error) {
	var m ResourceModel
	if err := r.db.Where("id = ? AND project_id = ?", resourceID, projectID).First(&m).Error; err != nil {
		return nil, err
	}
	return toResourceEntity(&m), nil
}

func (r *ResourceRepository) ListByProjectID(projectID string, resourceType *string) ([]*project.Resource, error) {
	var list []ResourceModel
	q := r.db.Where("project_id = ?", projectID)
	if resourceType != nil && *resourceType != "" {
		q = q.Where("type = ?", *resourceType)
	}
	if err := q.Order("created_at DESC").Find(&list).Error; err != nil {
		return nil, err
	}
	out := make([]*project.Resource, len(list))
	for i := range list {
		out[i] = toResourceEntity(&list[i])
	}
	return out, nil
}

func (r *ResourceRepository) Update(res *project.Resource) error {
	m := toResourceModel(res)
	return r.db.Model(&ResourceModel{}).Where("id = ? AND project_id = ?", res.ID, res.ProjectID).Updates(m).Error
}

func (r *ResourceRepository) Delete(projectID, resourceID string) error {
	res := r.db.Where("id = ? AND project_id = ?", resourceID, projectID).Delete(&ResourceModel{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gormErrNotFound
	}
	return nil
}

func toResourceModel(r *project.Resource) *ResourceModel {
	return &ResourceModel{
		ID:        r.ID,
		ProjectID: r.ProjectID,
		Type:      r.Type,
		Name:      r.Name,
		Content:   r.Content,
		URL:       r.URL,
		Size:      r.Size,
		CreatedAt: r.CreatedAt,
	}
}

func toResourceEntity(m *ResourceModel) *project.Resource {
	return &project.Resource{
		ID:        m.ID,
		ProjectID: m.ProjectID,
		Type:      m.Type,
		Name:      m.Name,
		Content:   m.Content,
		URL:       m.URL,
		Size:      m.Size,
		CreatedAt: m.CreatedAt,
	}
}
