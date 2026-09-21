package notification

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	TypeTaskAssigned      = "TASK_ASSIGNED"
	TypeTaskDueSoon       = "TASK_DUE_SOON"
	TypeTaskOverdue       = "TASK_OVERDUE"
	TypeTaskStatusChanged = "TASK_STATUS_CHANGED"
	TypeResourceOverload  = "RESOURCE_OVERLOAD"
	TypeProjectDueSoon    = "PROJECT_DUE_SOON"
	TypeMemberRemoved     = "MEMBER_REMOVED"
)

type Notification struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey"`
	UserID    uuid.UUID      `gorm:"type:uuid;not null;index"`
	Type      string         `gorm:"size:64;not null"`
	Title     string         `gorm:"size:255;not null"`
	Body      string         `gorm:"not null;default:''"`
	IsRead    bool           `gorm:"not null;default:false"`
	Metadata  map[string]any `gorm:"type:jsonb;serializer:json"`
	EntityID  uuid.UUID      `gorm:"type:uuid;not null"`
	Day       time.Time      `gorm:"type:date;not null"`
	CreatedAt time.Time
}

func (Notification) TableName() string { return "notifications" }

func (n *Notification) BeforeCreate(tx *gorm.DB) error {
	if n.ID == uuid.Nil {
		n.ID = uuid.New()
	}
	if n.Metadata == nil {
		n.Metadata = map[string]any{}
	}
	n.Day = DateUTC(n.Day)
	return nil
}

func DateUTC(t time.Time) time.Time {
	if t.IsZero() {
		t = time.Now().UTC()
	}
	y, m, d := t.UTC().Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

type Event struct {
	UserID   uuid.UUID
	Type     string
	Title    string
	Body     string
	EntityID uuid.UUID
	Day      time.Time
	Metadata map[string]any
}

type Public struct {
	ID        uuid.UUID      `json:"id"`
	Type      string         `json:"type"`
	Title     string         `json:"title"`
	Body      string         `json:"body"`
	IsRead    bool           `json:"isRead"`
	Metadata  map[string]any `json:"metadata"`
	EntityID  uuid.UUID      `json:"entityId"`
	CreatedAt time.Time      `json:"createdAt"`
}

func ToPublic(n Notification) Public {
	meta := n.Metadata
	if meta == nil {
		meta = map[string]any{}
	}
	return Public{
		ID: n.ID, Type: n.Type, Title: n.Title, Body: n.Body,
		IsRead: n.IsRead, Metadata: meta, EntityID: n.EntityID, CreatedAt: n.CreatedAt,
	}
}

type Page struct {
	Items      []Public `json:"items"`
	Page       int      `json:"page"`
	PageSize   int      `json:"pageSize"`
	TotalItems int64    `json:"totalItems"`
}

type ListFilter struct {
	UserID   uuid.UUID
	Unread   *bool
	Type     string
	Page     int
	PageSize int
}
