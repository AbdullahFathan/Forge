package resource

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"workspace/internal/project"
	"workspace/internal/user"
)

const (
	RoleLead   = project.RoleLead
	RoleMember = project.RoleMember
	RoleViewer = project.RoleViewer

	BandGreen  = "GREEN"
	BandYellow = "YELLOW"
	BandRed    = "RED"

	WarningOverAllocated = "OVER_ALLOCATED"

	matrixUserThreshold = 200
)

type Allocation struct {
	ID                uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID            uuid.UUID `gorm:"type:uuid;not null;index"`
	User              *user.User
	ProjectID         uuid.UUID `gorm:"type:uuid;not null;index"`
	Project           *project.Project
	AllocationPercent float64   `gorm:"type:numeric(5,2);not null"`
	StartDate         time.Time `gorm:"type:date;not null"`
	EndDate           time.Time `gorm:"type:date;not null"`
	Role              string    `gorm:"not null;size:32"`
	CreatedAt         time.Time
	UpdatedAt         time.Time
	DeletedAt         gorm.DeletedAt `gorm:"index"`
}

func (Allocation) TableName() string { return "resource_allocations" }

func (a *Allocation) BeforeCreate(tx *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return nil
}

type Holiday struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	Date      time.Time `gorm:"type:date;not null;uniqueIndex"`
	Name      string    `gorm:"not null;size:128"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (Holiday) TableName() string { return "holidays" }

func (h *Holiday) BeforeCreate(tx *gorm.DB) error {
	if h.ID == uuid.Nil {
		h.ID = uuid.New()
	}
	return nil
}

type Slice struct {
	UserID    uuid.UUID
	ProjectID uuid.UUID
	Percent   float64
	Start     time.Time
	End       time.Time
}
