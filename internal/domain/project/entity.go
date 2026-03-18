package project

import "time"

// Project 项目聚合根
type Project struct {
	ID          string
	UserID      string
	Name        string
	Description *string
	CoverImage  *string
	Status      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Session 会话实体（属于项目，一个项目下可有多个会话）
type Session struct {
	ID        string
	ProjectID string
	Title     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Message 消息实体（属于会话，会话属于项目）
type Message struct {
	ID         string
	ProjectID  string
	SessionID  string
	Role       string
	Content    string
	SkillID    *string
	Attachments map[string]interface{}
	CreatedAt  time.Time
}

// Resource 资源实体（属于项目聚合）
type Resource struct {
	ID        string
	ProjectID string
	Type      string
	Name      string
	Content   *string
	URL       *string
	Size      *string
	CreatedAt time.Time
}
