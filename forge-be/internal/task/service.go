package task

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/google/uuid"

	"workspace/internal/auditlog"
	"workspace/internal/project"
	"workspace/pkg/apperr"
	"workspace/pkg/authctx"
)

type Store interface {
	Create(*Task, []uuid.UUID) error
	Save(*Task, *[]uuid.UUID) error
	GetByID(uuid.UUID, bool) (*Task, error)
	List(ListFilter) ([]Task, int64, error)
	SoftDelete(uuid.UUID) error
	HasChildren(uuid.UUID) (bool, error)
	MaxPosition(projectID uuid.UUID, status string) (int, error)
	AddDependency(*Dependency) error
	DeleteDependency(taskID, depID uuid.UUID) error
	DependencyEdges(projectID uuid.UUID) (map[uuid.UUID][]uuid.UUID, error)
	UnfinishedPredecessors(taskID uuid.UUID) (bool, error)
	AddComment(*Comment) error
	ListComments(uuid.UUID) ([]Comment, error)
	IsAssignee(taskID, userID uuid.UUID) (bool, error)
}

type Projects interface {
	MustSee(actor authctx.Principal, projectID uuid.UUID) (*project.Project, string, error)
	MustManage(actor authctx.Principal, projectID uuid.UUID) (*project.Project, error)
	MemberRole(projectID, userID uuid.UUID) (bool, string, error)
}

type Service struct {
	repo  Store
	proj  Projects
	audit auditlog.Auditor
}

func NewService(repo Store, proj Projects, audit auditlog.Auditor) *Service {
	if audit == nil {
		audit = auditlog.Noop{}
	}
	return &Service{repo: repo, proj: proj, audit: audit}
}

type CreateInput struct {
	ProjectID      uuid.UUID
	ParentTaskID   *uuid.UUID
	Name           string
	Description    string
	Status         string
	Priority       string
	EstimatedHours *float64
	StartDate      *time.Time
	DueDate        *time.Time
	Labels         []string
	Position       *int
	AssigneeIDs    []uuid.UUID
}

type PatchInput struct {
	ParentTaskID   **uuid.UUID
	Name           *string
	Description    *string
	Status         *string
	Priority       *string
	EstimatedHours **float64
	StartDate      **time.Time
	DueDate        **time.Time
	Labels         *[]string
	Position       *int
	AssigneeIDs    *[]uuid.UUID
}

func (s *Service) List(actor authctx.Principal, f ListFilter) ([]Task, int64, error) {
	if _, _, err := s.proj.MustSee(actor, f.ProjectID); err != nil {
		return nil, 0, err
	}
	return s.repo.List(f)
}

func (s *Service) Get(actor authctx.Principal, id uuid.UUID, includeSubtasks bool) (*Task, error) {
	t, err := s.repo.GetByID(id, includeSubtasks)
	if err != nil {
		return nil, err
	}
	if _, _, err := s.proj.MustSee(actor, t.ProjectID); err != nil {
		return nil, err
	}
	return t, nil
}

func (s *Service) Create(ctx context.Context, actor authctx.Principal, ip string, in CreateInput) (*Task, error) {
	if _, err := s.proj.MustManage(actor, in.ProjectID); err != nil {
		return nil, err
	}
	if in.Status == "" {
		in.Status = StatusBacklog
	}
	if in.Priority == "" {
		in.Priority = PriorityMedium
	}
	if err := validateTaskEnums(in.Status, in.Priority); err != nil {
		return nil, err
	}
	if in.ParentTaskID != nil {
		parent, err := s.repo.GetByID(*in.ParentTaskID, false)
		if err != nil {
			return nil, apperr.ErrValidation.WithMessage("invalid parentTaskId")
		}
		if parent.ProjectID != in.ProjectID {
			return nil, apperr.ErrValidation.WithMessage("parent task must be in the same project")
		}
		if parent.ParentTaskID != nil {
			return nil, apperr.ErrValidation.WithMessage("subtask cannot have subtasks")
		}
	}
	if err := s.ensureAssigneesAreMembers(in.ProjectID, in.AssigneeIDs); err != nil {
		return nil, err
	}
	pos := 0
	if in.Position != nil {
		pos = *in.Position
	} else {
		max, err := s.repo.MaxPosition(in.ProjectID, in.Status)
		if err != nil {
			return nil, err
		}
		pos = max + 1
	}
	t := &Task{
		ProjectID: in.ProjectID, ParentTaskID: in.ParentTaskID,
		Name: strings.TrimSpace(in.Name), Description: in.Description,
		Status: in.Status, Priority: in.Priority, EstimatedHours: in.EstimatedHours,
		StartDate: in.StartDate, DueDate: in.DueDate, Labels: in.Labels, Position: pos,
	}
	if err := s.repo.Create(t, in.AssigneeIDs); err != nil {
		return nil, err
	}
	got, err := s.repo.GetByID(t.ID, false)
	if err != nil {
		return nil, err
	}
	_ = s.audit.Record(ctx, auditlog.Record{
		ActorID: actor.UserID, IP: ip, EntityType: "Task", EntityID: got.ID, Action: "CREATED", After: got,
	})
	return got, nil
}

func (s *Service) Patch(ctx context.Context, actor authctx.Principal, ip string, id uuid.UUID, in PatchInput) (*Task, error) {
	t, err := s.repo.GetByID(id, false)
	if err != nil {
		return nil, err
	}
	_, role, err := s.proj.MustSee(actor, t.ProjectID)
	if err != nil {
		return nil, err
	}
	manage := false
	if _, err := s.proj.MustManage(actor, t.ProjectID); err == nil {
		manage = true
	}
	assigned, err := s.repo.IsAssignee(t.ID, actor.UserID)
	if err != nil {
		return nil, err
	}
	if !manage {
		if project.IsViewer(role) || !assigned {
			return nil, apperr.ErrForbidden
		}
		if in.Status == nil && in.Position == nil {
			return nil, apperr.ErrForbidden.WithMessage("assignees may only update status or position")
		}
		if in.Name != nil || in.Description != nil || in.Priority != nil || in.AssigneeIDs != nil ||
			in.ParentTaskID != nil || in.Labels != nil || in.StartDate != nil || in.DueDate != nil || in.EstimatedHours != nil {
			return nil, apperr.ErrForbidden.WithMessage("assignees may only update status or position")
		}
	}
	before, _ := json.Marshal(t)
	action := "UPDATED"
	if in.Name != nil {
		t.Name = strings.TrimSpace(*in.Name)
	}
	if in.Description != nil {
		t.Description = *in.Description
	}
	if in.Status != nil {
		if err := ValidateTransition(t.Status, *in.Status); err != nil {
			return nil, err
		}
		if IsStartStatus(*in.Status) {
			blocked, err := s.repo.UnfinishedPredecessors(t.ID)
			if err != nil {
				return nil, err
			}
			if blocked {
				return nil, apperr.ErrValidation.WithMessage("cannot start task until predecessors are DONE")
			}
		}
		if t.Status != *in.Status {
			action = "STATUS_CHANGED"
		}
		t.Status = *in.Status
	}
	if in.Priority != nil {
		if err := validateTaskEnums(t.Status, *in.Priority); err != nil {
			return nil, err
		}
		t.Priority = *in.Priority
	}
	if in.EstimatedHours != nil {
		t.EstimatedHours = *in.EstimatedHours
	}
	if in.StartDate != nil {
		t.StartDate = *in.StartDate
	}
	if in.DueDate != nil {
		t.DueDate = *in.DueDate
	}
	if in.Labels != nil {
		t.Labels = *in.Labels
	}
	if in.Position != nil {
		t.Position = *in.Position
	}
	if in.ParentTaskID != nil {
		if err := s.applyParent(t, *in.ParentTaskID); err != nil {
			return nil, err
		}
	}
	if in.AssigneeIDs != nil {
		if err := s.ensureAssigneesAreMembers(t.ProjectID, *in.AssigneeIDs); err != nil {
			return nil, err
		}
	}
	if err := s.repo.Save(t, in.AssigneeIDs); err != nil {
		return nil, err
	}
	got, err := s.repo.GetByID(t.ID, false)
	if err != nil {
		return nil, err
	}
	_ = s.audit.Record(ctx, auditlog.Record{
		ActorID: actor.UserID, IP: ip, EntityType: "Task", EntityID: got.ID,
		Action: action, Before: json.RawMessage(before), After: got,
	})
	return got, nil
}

func (s *Service) Delete(ctx context.Context, actor authctx.Principal, ip string, id uuid.UUID) error {
	t, err := s.repo.GetByID(id, false)
	if err != nil {
		return err
	}
	if _, err := s.proj.MustManage(actor, t.ProjectID); err != nil {
		return err
	}
	if err := s.repo.SoftDelete(id); err != nil {
		return err
	}
	_ = s.audit.Record(ctx, auditlog.Record{
		ActorID: actor.UserID, IP: ip, EntityType: "Task", EntityID: id, Action: "DELETED", Before: t,
	})
	return nil
}

func (s *Service) AddDependency(ctx context.Context, actor authctx.Principal, ip string, taskID, dependsOn uuid.UUID) (*Task, error) {
	t, err := s.repo.GetByID(taskID, false)
	if err != nil {
		return nil, err
	}
	if _, err := s.proj.MustManage(actor, t.ProjectID); err != nil {
		return nil, err
	}
	if dependsOn == taskID {
		return nil, apperr.ErrValidation.WithMessage("task cannot depend on itself")
	}
	pred, err := s.repo.GetByID(dependsOn, false)
	if err != nil {
		return nil, apperr.ErrValidation.WithMessage("invalid dependsOnTaskId")
	}
	if pred.ProjectID != t.ProjectID {
		return nil, apperr.ErrValidation.WithMessage("dependency must be in the same project")
	}
	edges, err := s.repo.DependencyEdges(t.ProjectID)
	if err != nil {
		return nil, err
	}
	if WouldCycle(edges, taskID, dependsOn) {
		return nil, apperr.ErrValidation.WithMessage("dependency would create a cycle")
	}
	d := &Dependency{TaskID: taskID, DependsOnTaskID: dependsOn}
	if err := s.repo.AddDependency(d); err != nil {
		return nil, err
	}
	got, err := s.repo.GetByID(taskID, false)
	if err != nil {
		return nil, err
	}
	_ = s.audit.Record(ctx, auditlog.Record{
		ActorID: actor.UserID, IP: ip, EntityType: "Task", EntityID: taskID, Action: "UPDATED", After: got,
	})
	return got, nil
}

func (s *Service) RemoveDependency(ctx context.Context, actor authctx.Principal, ip string, taskID, depID uuid.UUID) error {
	t, err := s.repo.GetByID(taskID, false)
	if err != nil {
		return err
	}
	if _, err := s.proj.MustManage(actor, t.ProjectID); err != nil {
		return err
	}
	if err := s.repo.DeleteDependency(taskID, depID); err != nil {
		return err
	}
	_ = s.audit.Record(ctx, auditlog.Record{
		ActorID: actor.UserID, IP: ip, EntityType: "Task", EntityID: taskID, Action: "UPDATED",
	})
	return nil
}

func (s *Service) AddComment(ctx context.Context, actor authctx.Principal, ip string, taskID uuid.UUID, body string) (*Comment, error) {
	t, err := s.repo.GetByID(taskID, false)
	if err != nil {
		return nil, err
	}
	_, role, err := s.proj.MustSee(actor, t.ProjectID)
	if err != nil {
		return nil, err
	}
	_, manageErr := s.proj.MustManage(actor, t.ProjectID)
	assigned, err := s.repo.IsAssignee(taskID, actor.UserID)
	if err != nil {
		return nil, err
	}
	if manageErr != nil && (project.IsViewer(role) || !assigned) {
		return nil, apperr.ErrForbidden
	}
	body = strings.TrimSpace(body)
	if body == "" {
		return nil, apperr.ErrValidation.WithMessage("body is required")
	}
	c := &Comment{TaskID: taskID, UserID: actor.UserID, Body: body}
	if err := s.repo.AddComment(c); err != nil {
		return nil, err
	}
	rows, err := s.repo.ListComments(taskID)
	if err != nil {
		return nil, err
	}
	var created *Comment
	for i := range rows {
		if rows[i].ID == c.ID {
			created = &rows[i]
			break
		}
	}
	if created == nil {
		created = c
	}
	_ = s.audit.Record(ctx, auditlog.Record{
		ActorID: actor.UserID, IP: ip, EntityType: "Task", EntityID: taskID, Action: "UPDATED", After: created,
	})
	return created, nil
}

func (s *Service) ListComments(actor authctx.Principal, taskID uuid.UUID) ([]Comment, error) {
	t, err := s.repo.GetByID(taskID, false)
	if err != nil {
		return nil, err
	}
	if _, _, err := s.proj.MustSee(actor, t.ProjectID); err != nil {
		return nil, err
	}
	return s.repo.ListComments(taskID)
}

func (s *Service) applyParent(t *Task, parentID *uuid.UUID) error {
	if parentID == nil {
		t.ParentTaskID = nil
		return nil
	}
	if *parentID == t.ID {
		return apperr.ErrValidation.WithMessage("task cannot be its own parent")
	}
	parent, err := s.repo.GetByID(*parentID, false)
	if err != nil {
		return apperr.ErrValidation.WithMessage("invalid parentTaskId")
	}
	if parent.ProjectID != t.ProjectID {
		return apperr.ErrValidation.WithMessage("parent task must be in the same project")
	}
	if parent.ParentTaskID != nil {
		return apperr.ErrValidation.WithMessage("subtask cannot have subtasks")
	}
	hasKids, err := s.repo.HasChildren(t.ID)
	if err != nil {
		return err
	}
	if hasKids {
		return apperr.ErrValidation.WithMessage("cannot nest a parent task under another task")
	}
	t.ParentTaskID = parentID
	return nil
}

func (s *Service) ensureAssigneesAreMembers(projectID uuid.UUID, ids []uuid.UUID) error {
	for _, id := range ids {
		ok, _, err := s.proj.MemberRole(projectID, id)
		if err != nil {
			return err
		}
		if !ok {
			return apperr.ErrValidation.WithMessage("assignees must be project members")
		}
	}
	return nil
}

func validateTaskEnums(status, priority string) error {
	switch status {
	case StatusBacklog, StatusTodo, StatusInProgress, StatusInReview, StatusDone, StatusBlocked:
	default:
		return apperr.ErrValidation.WithMessage("invalid status")
	}
	switch priority {
	case PriorityLow, PriorityMedium, PriorityHigh, PriorityCritical:
	default:
		return apperr.ErrValidation.WithMessage("invalid priority")
	}
	return nil
}
