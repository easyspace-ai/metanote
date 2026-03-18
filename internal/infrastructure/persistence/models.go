package persistence

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

// UserModel GORM 用户表模型
type UserModel struct {
	ID               string    `gorm:"primaryKey"`
	Username         string    `gorm:"uniqueIndex;not null"`
	Email            string    `gorm:"uniqueIndex;not null"`
	HashedPassword   string    `gorm:"not null"`
	SubscriptionPlan string    `gorm:"not null;default:free"`
	CreditsBalance   int       `gorm:"not null;default:1000"`
	CreditsUsed      int       `gorm:"not null;default:0"`
	CreatedAt        time.Time `gorm:"not null"`
}

func (UserModel) TableName() string { return "users" }

// ProjectModel GORM 项目表模型
type ProjectModel struct {
	ID          string     `gorm:"primaryKey"`
	UserID      string     `gorm:"index;not null"`
	Name        string     `gorm:"not null"`
	Description *string    `gorm:"type:text"`
	CoverImage  *string    `gorm:""`
	Status      string     `gorm:"not null;default:active"`
	CreatedAt   time.Time  `gorm:"not null"`
	UpdatedAt   time.Time  `gorm:"not null"`
}

func (ProjectModel) TableName() string { return "projects" }

// SessionModel GORM 会话表模型
type SessionModel struct {
	ID        string    `gorm:"primaryKey"`
	ProjectID string    `gorm:"index;not null"`
	Title     string    `gorm:"not null"`
	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`
}

func (SessionModel) TableName() string { return "sessions" }

// MessageModel GORM 消息表模型
type MessageModel struct {
	ID          string    `gorm:"primaryKey"`
	ProjectID   string    `gorm:"index;not null"`
	SessionID   string    `gorm:"index;not null"`
	Role        string    `gorm:"not null"`
	Content     string    `gorm:"type:text;not null"`
	SkillID     *string   `gorm:""`
	Attachments string    `gorm:"type:text"`
	CreatedAt   time.Time `gorm:"not null"`
}

func (MessageModel) TableName() string { return "messages" }

// ResourceModel GORM 资源表模型
type ResourceModel struct {
	ID        string     `gorm:"primaryKey"`
	ProjectID string     `gorm:"index;not null"`
	Type      string     `gorm:"not null"`
	Name      string     `gorm:"not null"`
	Content   *string    `gorm:"type:text"`
	URL       *string    `gorm:""`
	Size      *string    `gorm:""`
	CreatedAt time.Time  `gorm:"not null"`
}

func (ResourceModel) TableName() string { return "resources" }

// SkillModel GORM 技能表模型
type SkillModel struct {
	ID            string    `gorm:"primaryKey"`
	Name          string    `gorm:"not null"`
	Description   *string   `gorm:"type:text"`
	Icon          *string   `gorm:""`
	Category      string    `gorm:"not null;default:other"`
	Author        *string   `gorm:""`
	UsersCount    int       `gorm:"not null;default:0"`
	Rating        float64   `gorm:"not null;default:0"`
	Tags          JSONSlice `gorm:"type:text"`
	SystemPrompt  *string   `gorm:"type:text"`
	IsInstalled   bool      `gorm:"not null;default:false"`
	IsPersonal    bool      `gorm:"not null;default:false"`
	IsRecommended bool      `gorm:"not null;default:false"`
	CreatedAt     time.Time `gorm:"not null"`
	UpdatedAt     time.Time `gorm:"not null"`
}

func (SkillModel) TableName() string { return "skills" }

// JSONSlice 用于 GORM 读写 []string 到 json (SQLite)
type JSONSlice []string

func (s JSONSlice) Value() (driver.Value, error) {
	if s == nil {
		return nil, nil
	}
	return json.Marshal(s)
}

func (s *JSONSlice) Scan(value interface{}) error {
	if value == nil {
		*s = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		// SQLite 有时返回 string
		str, ok := value.(string)
		if !ok {
			return errors.New("invalid type for JSONSlice")
		}
		bytes = []byte(str)
	}
	return json.Unmarshal(bytes, s)
}
