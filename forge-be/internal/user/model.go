package user

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"workspace/internal/department"
	"workspace/internal/rbac"
)

type User struct {
	ID                  uuid.UUID  `gorm:"type:uuid;primaryKey"`
	Name                string     `gorm:"not null;size:128"`
	Email               string     `gorm:"not null;size:255;index"`
	PasswordHash        string     `gorm:"not null"`
	RoleID              uuid.UUID  `gorm:"type:uuid;not null;index"`
	Role                rbac.Role  `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
	DepartmentID        *uuid.UUID `gorm:"type:uuid;index"`
	Department          *department.Department
	CapacityHoursPerDay int     `gorm:"not null;default:8"`
	IsActive            bool    `gorm:"not null;default:true"`
	Skills              []Skill `gorm:"foreignKey:UserID"`
	CreatedAt           time.Time
	UpdatedAt           time.Time
	DeletedAt           gorm.DeletedAt `gorm:"index"`
}

type Skill struct {
	UserID uuid.UUID `gorm:"type:uuid;primaryKey"`
	Skill  string    `gorm:"primaryKey;size:64"`
}

func (Skill) TableName() string { return "user_skills" }

func SkillNames(u *User) []string {
	if u == nil || len(u.Skills) == 0 {
		return []string{}
	}
	out := make([]string, 0, len(u.Skills))
	for _, s := range u.Skills {
		out = append(out, s.Skill)
	}
	return out
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	if u.CapacityHoursPerDay == 0 {
		u.CapacityHoursPerDay = 8
	}
	return nil
}
