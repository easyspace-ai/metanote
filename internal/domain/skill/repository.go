package skill

// Repository 技能仓储接口
type Repository interface {
	Create(s *Skill) error
	GetByID(id string) (*Skill, error)
	SetInstalled(id string, installed bool) error
	ListAll() ([]*Skill, error)
	ListInstalled() ([]*Skill, error)
	ListRecommended(limit int) ([]*Skill, error)
}
