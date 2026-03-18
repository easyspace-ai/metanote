package skill

// Skill 技能实体（读多写少，可视为独立聚合）
type Skill struct {
	ID            string
	Name          string
	Description   *string
	Icon          *string
	Category      string
	Author        *string
	UsersCount    int
	Rating        float64
	Tags          []string
	SystemPrompt  *string
	IsInstalled   bool
	IsPersonal    bool
	IsRecommended bool
}
