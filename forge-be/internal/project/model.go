package project

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"workspace/internal/department"
	"workspace/internal/user"
)

const (
	StatusDraft     = "DRAFT"
	StatusActive    = "ACTIVE"
	StatusOnHold    = "ON_HOLD"
	StatusCompleted = "COMPLETED"
	StatusArchived  = "ARCHIVED"

	PriorityLow      = "LOW"
	PriorityMedium   = "MEDIUM"
	PriorityHigh     = "HIGH"
	PriorityCritical = "CRITICAL"

	RoleLead   = "LEAD"
	RoleMember = "MEMBER"
	RoleViewer = "VIEWER"
)

type Project struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey"`
	Name          string    `gorm:"not null;size:200"`
	Description   string    `gorm:"not null;default:''"`
	Status        string    `gorm:"not null;size:32"`
	Priority      string    `gorm:"not null;size:32"`
	StartDate     time.Time `gorm:"type:date;not null"`
	TargetEndDate time.Time `gorm:"type:date;not null"`
	OwnerID       uuid.UUID `gorm:"type:uuid;not null;index"`
	Owner         *user.User
	DepartmentID  *uuid.UUID `gorm:"type:uuid;index"`
	Department    *department.Department
	Tags          []string `gorm:"type:jsonb;serializer:json"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     gorm.DeletedAt `gorm:"index"`
	Members       []Member
}

func (Project) TableName() string { return "projects" }

func (p *Project) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	if p.Tags == nil {
		p.Tags = []string{}
	}
	return nil
}

type Member struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	ProjectID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:ux_project_members_project_user"`
	Project   *Project
	UserID    uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:ux_project_members_project_user"`
	User      *user.User
	Role      string `gorm:"not null;size:32"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (Member) TableName() string { return "project_members" }

func (m *Member) BeforeCreate(tx *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}
