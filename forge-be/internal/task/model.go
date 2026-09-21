package task

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"workspace/internal/user"
)

const (
	StatusBacklog    = "BACKLOG"
	StatusTodo       = "TODO"
	StatusInProgress = "IN_PROGRESS"
	StatusInReview   = "IN_REVIEW"
	StatusDone       = "DONE"
	StatusBlocked    = "BLOCKED"

	PriorityLow      = "LOW"
	PriorityMedium   = "MEDIUM"
	PriorityHigh     = "HIGH"
	PriorityCritical = "CRITICAL"
)

type Task struct {
	ID             uuid.UUID  `gorm:"type:uuid;primaryKey"`
	ProjectID      uuid.UUID  `gorm:"type:uuid;not null;index"`
	ParentTaskID   *uuid.UUID `gorm:"type:uuid;index"`
	Parent         *Task      `gorm:"foreignKey:ParentTaskID"`
	Name           string     `gorm:"not null;size:200"`
	Description    string     `gorm:"not null;default:''"`
	Status         string     `gorm:"not null;size:32"`
	Priority       string     `gorm:"not null;size:32"`
	EstimatedHours *float64
	StartDate      *time.Time `gorm:"type:date"`
	DueDate        *time.Time `gorm:"type:date"`
	Labels         []string   `gorm:"type:jsonb;serializer:json"`
	Position       int        `gorm:"not null;default:0"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      gorm.DeletedAt `gorm:"index"`
	Assignees      []user.User    `gorm:"many2many:task_assignees;"`
	DependsOn      []Dependency   `gorm:"foreignKey:TaskID"`
	Dependents     []Dependency   `gorm:"foreignKey:DependsOnTaskID"`
	Subtasks       []Task         `gorm:"foreignKey:ParentTaskID"`
}

func (Task) TableName() string { return "tasks" }

func (t *Task) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	if t.Labels == nil {
		t.Labels = []string{}
	}
	return nil
}

type Dependency struct {
	ID              uuid.UUID `gorm:"type:uuid;primaryKey"`
	TaskID          uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:ux_task_deps"`
	DependsOnTaskID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:ux_task_deps"`
	DependsOn       *Task     `gorm:"foreignKey:DependsOnTaskID"`
}

func (Dependency) TableName() string { return "task_dependencies" }

func (d *Dependency) BeforeCreate(tx *gorm.DB) error {
	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}
	return nil
}

type Comment struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	TaskID    uuid.UUID `gorm:"type:uuid;not null;index"`
	UserID    uuid.UUID `gorm:"type:uuid;not null"`
	User      *user.User
	Body      string `gorm:"not null"`
	CreatedAt time.Time
}

func (Comment) TableName() string { return "task_comments" }

func (c *Comment) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}
