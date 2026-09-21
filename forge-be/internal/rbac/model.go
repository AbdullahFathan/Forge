package rbac

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Permission struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	Code      string    `gorm:"uniqueIndex;not null;size:64"`
	Name      string    `gorm:"not null;size:128"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (p *Permission) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}

type Role struct {
	ID          uuid.UUID    `gorm:"type:uuid;primaryKey"`
	Code        string       `gorm:"uniqueIndex;not null;size:64"`
	Name        string       `gorm:"not null;size:128"`
	IsSystem    bool         `gorm:"not null;default:true"`
	Permissions []Permission `gorm:"many2many:role_permissions;"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (r *Role) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return nil
}
