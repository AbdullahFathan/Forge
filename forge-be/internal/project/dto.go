package project

import (
	"time"

	"github.com/google/uuid"
)

type Public struct {
	ID                uuid.UUID  `json:"id"`
	Name              string     `json:"name"`
	Description       string     `json:"description"`
	Status            string     `json:"status"`
	Priority          string     `json:"priority"`
	StartDate         string     `json:"startDate"`
	TargetEndDate     string     `json:"targetEndDate"`
	OwnerID           uuid.UUID  `json:"ownerId"`
	OwnerName         string     `json:"ownerName,omitempty"`
	DepartmentID      *uuid.UUID `json:"departmentId"`
	DepartmentName    *string    `json:"departmentName,omitempty"`
	Tags              []string   `json:"tags"`
	CompletionPercent float64    `json:"completionPercent"`
	CreatedAt         time.Time  `json:"createdAt"`
	UpdatedAt         time.Time  `json:"updatedAt"`
}

type TaskCounts struct {
	Backlog    int64 `json:"BACKLOG"`
	Todo       int64 `json:"TODO"`
	InProgress int64 `json:"IN_PROGRESS"`
	InReview   int64 `json:"IN_REVIEW"`
	Done       int64 `json:"DONE"`
	Blocked    int64 `json:"BLOCKED"`
	Total      int64 `json:"total"`
}

type Summary struct {
	Public
	TaskCounts  TaskCounts `json:"taskCounts"`
	MemberCount int64      `json:"memberCount"`
	Activity    []any      `json:"activity"` // TODO Phase 4: project audit activity
}

type MemberPublic struct {
	UserID         uuid.UUID  `json:"userId"`
	Name           string     `json:"name"`
	Email          string     `json:"email,omitempty"`
	DepartmentID   *uuid.UUID `json:"departmentId"`
	DepartmentName *string    `json:"departmentName,omitempty"`
	ProjectRole    string     `json:"projectRole"`
}

type Page struct {
	Items      []Public `json:"items"`
	Page       int      `json:"page"`
	PageSize   int      `json:"pageSize"`
	TotalItems int64    `json:"totalItems"`
}

type MemberPage struct {
	Items      []MemberPublic `json:"items"`
	Page       int            `json:"page"`
	PageSize   int            `json:"pageSize"`
	TotalItems int64          `json:"totalItems"`
}

func dateStr(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format("2006-01-02")
}

func ToPublic(p *Project, completion float64) Public {
	out := Public{
		ID:                p.ID,
		Name:              p.Name,
		Description:       p.Description,
		Status:            p.Status,
		Priority:          p.Priority,
		StartDate:         dateStr(p.StartDate),
		TargetEndDate:     dateStr(p.TargetEndDate),
		OwnerID:           p.OwnerID,
		DepartmentID:      p.DepartmentID,
		Tags:              p.Tags,
		CompletionPercent: completion,
		CreatedAt:         p.CreatedAt,
		UpdatedAt:         p.UpdatedAt,
	}
	if out.Tags == nil {
		out.Tags = []string{}
	}
	if p.Owner != nil {
		out.OwnerName = p.Owner.Name
	}
	if p.Department != nil {
		out.DepartmentName = &p.Department.Name
	}
	return out
}

func ToMemberPublic(m *Member) MemberPublic {
	out := MemberPublic{
		UserID:      m.UserID,
		ProjectRole: m.Role,
	}
	if m.User != nil {
		out.Name = m.User.Name
		out.Email = m.User.Email
		out.DepartmentID = m.User.DepartmentID
		if m.User.Department != nil {
			out.DepartmentName = &m.User.Department.Name
		}
	}
	return out
}
