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
	CapacityHoursPerDay int  `gorm:"not null;default:8"`
	IsActive            bool `gorm:"not null;default:true"`
	CreatedAt           time.Time
	UpdatedAt           time.Time
	DeletedAt           gorm.DeletedAt `gorm:"index"`
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
