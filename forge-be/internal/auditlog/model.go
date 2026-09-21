package auditlog

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Log struct {
	ID         uuid.UUID       `gorm:"type:uuid;primaryKey"`
	UserID     uuid.UUID       `gorm:"type:uuid;not null;index"`
	IPAddress  string          `gorm:"size:64;not null;default:''"`
	EntityType string          `gorm:"size:64;not null"`
	EntityID   uuid.UUID       `gorm:"type:uuid;not null"`
	ProjectID  *uuid.UUID      `gorm:"type:uuid;index"`
	Action     string          `gorm:"size:32;not null"`
	Before     json.RawMessage `gorm:"type:jsonb"`
	After      json.RawMessage `gorm:"type:jsonb"`
	CreatedAt  time.Time
}

func (Log) TableName() string { return "audit_logs" }

type Public struct {
	ID         uuid.UUID       `json:"id"`
	UserID     uuid.UUID       `json:"userId"`
	IPAddress  string          `json:"ipAddress"`
	EntityType string          `json:"entityType"`
	EntityID   uuid.UUID       `json:"entityId"`
	ProjectID  *uuid.UUID      `json:"projectId,omitempty"`
	Action     string          `json:"action"`
	Before     json.RawMessage `json:"before,omitempty"`
	After      json.RawMessage `json:"after,omitempty"`
	CreatedAt  time.Time       `json:"createdAt"`
}

func ToPublic(l Log) Public {
	return Public{
		ID: l.ID, UserID: l.UserID, IPAddress: l.IPAddress,
		EntityType: l.EntityType, EntityID: l.EntityID, ProjectID: l.ProjectID,
		Action: l.Action, Before: l.Before, After: l.After, CreatedAt: l.CreatedAt,
	}
}

type Page struct {
	Items      []Public `json:"items"`
	Page       int      `json:"page"`
	PageSize   int      `json:"pageSize"`
	TotalItems int64    `json:"totalItems"`
}

type ListFilter struct {
	Page       int
	PageSize   int
	From       *time.Time
	To         *time.Time
	UserID     *uuid.UUID
	EntityType string
	Action     string
	ProjectID  *uuid.UUID
}
