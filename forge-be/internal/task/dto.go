package task

import (
	"time"

	"github.com/google/uuid"
)

type UserBrief struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

type DepRef struct {
	ID     uuid.UUID `json:"id"`
	TaskID uuid.UUID `json:"taskId"`
	Name   string    `json:"name,omitempty"`
	Status string    `json:"status,omitempty"`
}

type Public struct {
	ID             uuid.UUID   `json:"id"`
	ProjectID      uuid.UUID   `json:"projectId"`
	ParentTaskID   *uuid.UUID  `json:"parentTaskId"`
	Name           string      `json:"name"`
	Description    string      `json:"description"`
	Status         string      `json:"status"`
	Priority       string      `json:"priority"`
	EstimatedHours *float64    `json:"estimatedHours"`
	StartDate      *string     `json:"startDate"`
	DueDate        *string     `json:"dueDate"`
	Labels         []string    `json:"labels"`
	Position       int         `json:"position"`
	Assignees      []UserBrief `json:"assignees"`
	DependsOn      []DepRef    `json:"dependsOn"`
	Dependents     []DepRef    `json:"dependents"`
	Warnings       []string    `json:"warnings"`
	Subtasks       []Public    `json:"subtasks,omitempty"`
	CreatedAt      time.Time   `json:"createdAt"`
	UpdatedAt      time.Time   `json:"updatedAt"`
}

type CommentPublic struct {
	ID        uuid.UUID `json:"id"`
	TaskID    uuid.UUID `json:"taskId"`
	UserID    uuid.UUID `json:"userId"`
	UserName  string    `json:"userName,omitempty"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"createdAt"`
}

type Page struct {
	Items      []Public `json:"items"`
	Page       int      `json:"page"`
	PageSize   int      `json:"pageSize"`
	TotalItems int64    `json:"totalItems"`
}

func datePtr(t *time.Time) *string {
	if t == nil || t.IsZero() {
		return nil
	}
	s := t.UTC().Format("2006-01-02")
	return &s
}

func ToPublic(t *Task, includeSubtasks bool) Public {
	out := Public{
		ID: t.ID, ProjectID: t.ProjectID, ParentTaskID: t.ParentTaskID,
		Name: t.Name, Description: t.Description, Status: t.Status, Priority: t.Priority,
		EstimatedHours: t.EstimatedHours, StartDate: datePtr(t.StartDate), DueDate: datePtr(t.DueDate),
		Labels: t.Labels, Position: t.Position, CreatedAt: t.CreatedAt, UpdatedAt: t.UpdatedAt,
		Assignees: []UserBrief{}, DependsOn: []DepRef{}, Dependents: []DepRef{}, Warnings: []string{},
	}
	if out.Labels == nil {
		out.Labels = []string{}
	}
	for _, u := range t.Assignees {
		out.Assignees = append(out.Assignees, UserBrief{ID: u.ID, Name: u.Name})
	}
	for _, d := range t.DependsOn {
		ref := DepRef{ID: d.ID, TaskID: d.DependsOnTaskID}
		if d.DependsOn != nil {
			ref.Name = d.DependsOn.Name
			ref.Status = d.DependsOn.Status
			if IsStartStatus(t.Status) && d.DependsOn.Status != StatusDone {
				out.Warnings = append(out.Warnings, "predecessor "+d.DependsOnTaskID.String()+" is not DONE")
			}
		}
		out.DependsOn = append(out.DependsOn, ref)
	}
	for _, d := range t.Dependents {
		out.Dependents = append(out.Dependents, DepRef{ID: d.ID, TaskID: d.TaskID})
	}
	if includeSubtasks {
		out.Subtasks = make([]Public, 0, len(t.Subtasks))
		for i := range t.Subtasks {
			out.Subtasks = append(out.Subtasks, ToPublic(&t.Subtasks[i], false))
		}
	}
	return out
}

func ToCommentPublic(c *Comment) CommentPublic {
	out := CommentPublic{ID: c.ID, TaskID: c.TaskID, UserID: c.UserID, Body: c.Body, CreatedAt: c.CreatedAt}
	if c.User != nil {
		out.UserName = c.User.Name
	}
	return out
}
