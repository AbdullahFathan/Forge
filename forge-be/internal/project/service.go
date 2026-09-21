package project

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"

	"workspace/internal/auditlog"
	"workspace/internal/department"
	"workspace/internal/rbac/perm"
	"workspace/internal/user"
	"workspace/pkg/apperr"
	"workspace/pkg/authctx"
)

type Store interface {
	Create(*Project) error
	Save(*Project) error
	GetByID(uuid.UUID) (*Project, error)
	List(ListFilter) ([]Project, int64, error)
	Archive(uuid.UUID) error
	GetMember(projectID, userID uuid.UUID) (*Member, error)
	ListMembers(projectID uuid.UUID, page, pageSize int) ([]Member, int64, error)
	CountMembers(uuid.UUID) (int64, error)
	AddMember(*Member) error
	SaveMember(*Member) error
	DeleteMember(projectID, userID uuid.UUID) error
	IsMember(projectID, userID uuid.UUID) (bool, string, error)
}

type UserFinder interface {
	GetByID(uuid.UUID) (*user.User, error)
}

type DeptFinder interface {
	GetByID(uuid.UUID) (*department.Department, error)
}

type TaskStats interface {
	Completion(projectID uuid.UUID) (percent float64, counts TaskCounts, err error)
}

type Service struct {
	repo  Store
	users UserFinder
	depts DeptFinder
	stats TaskStats
	audit auditlog.Auditor
}

func NewService(repo Store, users UserFinder, depts DeptFinder, stats TaskStats, audit auditlog.Auditor) *Service {
	if audit == nil {
		audit = auditlog.Noop{}
	}
	return &Service{repo: repo, users: users, depts: depts, stats: stats, audit: audit}
}

type CreateInput struct {
	Name          string
	Description   string
	Status        string
	Priority      string
	StartDate     time.Time
	TargetEndDate time.Time
	OwnerID       uuid.UUID
	DepartmentID  *uuid.UUID
	Tags          []string
}

type PatchInput struct {
	Name          *string
	Description   *string
	Status        *string
	Priority      *string
	StartDate     *time.Time
	TargetEndDate *time.Time
	OwnerID       *uuid.UUID
	DepartmentID  **uuid.UUID
	Tags          *[]string
}

func (s *Service) List(actor authctx.Principal, f ListFilter) ([]Project, []float64, int64, error) {
	f.UserID = actor.UserID
	f.ScopeAll = authctx.HasPermission(actor, perm.ProjectReadAll)
	rows, total, err := s.repo.List(f)
	if err != nil {
		return nil, nil, 0, err
	}
	percents := make([]float64, len(rows))
	for i := range rows {
		percents[i], _, _ = s.completion(rows[i].ID)
	}
	return rows, percents, total, nil
}

func (s *Service) Get(actor authctx.Principal, id uuid.UUID) (*Project, Summary, error) {
	p, memberRole, err := s.loadVisible(actor, id)
	if err != nil {
		return nil, Summary{}, err
	}
	_ = memberRole
	sum, err := s.summary(p)
	return p, sum, err
}

func (s *Service) Create(ctx context.Context, actor authctx.Principal, ip string, in CreateInput) (*Project, error) {
	if in.Status == "" {
		in.Status = StatusDraft
	}
	if in.Priority == "" {
		in.Priority = PriorityMedium
	}
	if err := validateStatusPriority(in.Status, in.Priority); err != nil {
		return nil, err
	}
	if in.TargetEndDate.Before(in.StartDate) {
		return nil, apperr.ErrValidation.WithMessage("targetEndDate must be on or after startDate")
	}
	if _, err := s.users.GetByID(in.OwnerID); err != nil {
		return nil, apperr.ErrValidation.WithMessage("invalid ownerId")
	}
	if in.DepartmentID != nil {
		if _, err := s.depts.GetByID(*in.DepartmentID); err != nil {
			return nil, apperr.ErrValidation.WithMessage("invalid departmentId")
		}
	}
	p := &Project{
		Name:          strings.TrimSpace(in.Name),
		Description:   in.Description,
		Status:        in.Status,
		Priority:      in.Priority,
		StartDate:     in.StartDate,
		TargetEndDate: in.TargetEndDate,
		OwnerID:       in.OwnerID,
		DepartmentID:  in.DepartmentID,
		Tags:          in.Tags,
	}
	if err := s.repo.Create(p); err != nil {
		return nil, err
	}
	created, err := s.repo.GetByID(p.ID)
	if err != nil {
		return nil, err
	}
	_ = s.audit.Record(ctx, auditlog.Record{
		ActorID: actor.UserID, IP: ip, EntityType: "Project", EntityID: created.ID,
		Action: "CREATED", After: created,
	})
	return created, nil
}

func (s *Service) Patch(ctx context.Context, actor authctx.Principal, ip string, id uuid.UUID, in PatchInput) (*Project, error) {
	p, role, err := s.loadVisible(actor, id)
	if err != nil {
		return nil, err
	}
	if !CanManage(actor, p, role) {
		return nil, apperr.ErrForbidden
	}
	before := *p
	if in.Name != nil {
		p.Name = strings.TrimSpace(*in.Name)
	}
	if in.Description != nil {
		p.Description = *in.Description
	}
	if in.Status != nil {
		if err := ValidateTransition(p.Status, *in.Status); err != nil {
			return nil, err
		}
		p.Status = *in.Status
	}
	if in.Priority != nil {
		if err := validateStatusPriority(p.Status, *in.Priority); err != nil {
			return nil, err
		}
		p.Priority = *in.Priority
	}
	if in.StartDate != nil {
		p.StartDate = *in.StartDate
	}
	if in.TargetEndDate != nil {
		p.TargetEndDate = *in.TargetEndDate
	}
	if p.TargetEndDate.Before(p.StartDate) {
		return nil, apperr.ErrValidation.WithMessage("targetEndDate must be on or after startDate")
	}
	if in.OwnerID != nil && *in.OwnerID != p.OwnerID {
		if _, err := s.users.GetByID(*in.OwnerID); err != nil {
			return nil, apperr.ErrValidation.WithMessage("invalid ownerId")
		}
		ok, _, err := s.repo.IsMember(p.ID, *in.OwnerID)
		if err != nil {
			return nil, err
		}
		if !ok {
			if err := s.repo.AddMember(&Member{ProjectID: p.ID, UserID: *in.OwnerID, Role: RoleLead}); err != nil {
				return nil, conflictOr(err)
			}
		}
		p.OwnerID = *in.OwnerID
	}
	if in.DepartmentID != nil {
		if *in.DepartmentID != nil {
			if _, err := s.depts.GetByID(**in.DepartmentID); err != nil {
				return nil, apperr.ErrValidation.WithMessage("invalid departmentId")
			}
		}
		p.DepartmentID = *in.DepartmentID
	}
	if in.Tags != nil {
		p.Tags = *in.Tags
	}
	if err := s.repo.Save(p); err != nil {
		return nil, err
	}
	updated, err := s.repo.GetByID(p.ID)
	if err != nil {
		return nil, err
	}
	_ = s.audit.Record(ctx, auditlog.Record{
		ActorID: actor.UserID, IP: ip, EntityType: "Project", EntityID: updated.ID,
		Action: "UPDATED", Before: before, After: updated,
	})
	return updated, nil
}

func (s *Service) Delete(ctx context.Context, actor authctx.Principal, ip string, id uuid.UUID) error {
	p, role, err := s.loadVisible(actor, id)
	if err != nil {
		return err
	}
	if !CanManage(actor, p, role) {
		return apperr.ErrForbidden
	}
	if err := ValidateTransition(p.Status, StatusArchived); err != nil {
		return err
	}
	if err := s.repo.Archive(id); err != nil {
		return err
	}
	_ = s.audit.Record(ctx, auditlog.Record{
		ActorID: actor.UserID, IP: ip, EntityType: "Project", EntityID: id,
		Action: "DELETED", Before: p,
	})
	return nil
}

func (s *Service) ListMembers(actor authctx.Principal, projectID uuid.UUID, page, pageSize int) ([]Member, int64, error) {
	if _, _, err := s.loadVisible(actor, projectID); err != nil {
		return nil, 0, err
	}
	return s.repo.ListMembers(projectID, page, pageSize)
}

func (s *Service) AddMember(ctx context.Context, actor authctx.Principal, ip string, projectID, userID uuid.UUID, role string) (*Member, error) {
	p, memberRole, err := s.loadVisible(actor, projectID)
	if err != nil {
		return nil, err
	}
	if !CanManage(actor, p, memberRole) {
		return nil, apperr.ErrForbidden
	}
	if role == "" {
		role = RoleMember
	}
	if role != RoleLead && role != RoleMember && role != RoleViewer {
		return nil, apperr.ErrValidation.WithMessage("invalid project role")
	}
	if _, err := s.users.GetByID(userID); err != nil {
		return nil, apperr.ErrValidation.WithMessage("invalid userId")
	}
	m := &Member{ProjectID: projectID, UserID: userID, Role: role}
	if err := s.repo.AddMember(m); err != nil {
		return nil, conflictOr(err)
	}
	got, err := s.repo.GetMember(projectID, userID)
	if err != nil {
		return nil, err
	}
	_ = s.audit.Record(ctx, auditlog.Record{
		ActorID: actor.UserID, IP: ip, EntityType: "ProjectMember", EntityID: got.ID,
		Action: "CREATED", After: got,
	})
	return got, nil
}

func (s *Service) RemoveMember(ctx context.Context, actor authctx.Principal, ip string, projectID, userID uuid.UUID) error {
	p, memberRole, err := s.loadVisible(actor, projectID)
	if err != nil {
		return err
	}
	if !CanManage(actor, p, memberRole) {
		return apperr.ErrForbidden
	}
	if userID == p.OwnerID {
		return apperr.ErrValidation.WithMessage("cannot remove project owner without transferring ownership")
	}
	m, err := s.repo.GetMember(projectID, userID)
	if err != nil {
		return err
	}
	if err := s.repo.DeleteMember(projectID, userID); err != nil {
		return err
	}
	_ = s.audit.Record(ctx, auditlog.Record{
		ActorID: actor.UserID, IP: ip, EntityType: "ProjectMember", EntityID: m.ID,
		Action: "DELETED", Before: m,
	})
	return nil
}

func (s *Service) MemberRole(projectID, userID uuid.UUID) (bool, string, error) {
	return s.repo.IsMember(projectID, userID)
}

func (s *Service) MustManage(actor authctx.Principal, projectID uuid.UUID) (*Project, error) {
	p, role, err := s.loadVisible(actor, projectID)
	if err != nil {
		return nil, err
	}
	if !CanManage(actor, p, role) {
		return nil, apperr.ErrForbidden
	}
	return p, nil
}

func (s *Service) MustSee(actor authctx.Principal, projectID uuid.UUID) (*Project, string, error) {
	return s.loadVisible(actor, projectID)
}

func (s *Service) loadVisible(actor authctx.Principal, id uuid.UUID) (*Project, string, error) {
	p, err := s.repo.GetByID(id)
	if err != nil {
		return nil, "", err
	}
	ok, role, err := s.repo.IsMember(id, actor.UserID)
	if err != nil {
		return nil, "", err
	}
	if p.OwnerID == actor.UserID && !ok {
		ok = true
		role = RoleLead
	}
	if !CanSee(actor, p, role, ok) {
		return nil, "", apperr.ErrForbidden
	}
	return p, role, nil
}

func (s *Service) summary(p *Project) (Summary, error) {
	pct, counts, err := s.completion(p.ID)
	if err != nil {
		return Summary{}, err
	}
	n, err := s.repo.CountMembers(p.ID)
	if err != nil {
		return Summary{}, err
	}
	return Summary{
		Public:      ToPublic(p, pct),
		TaskCounts:  counts,
		MemberCount: n,
		Activity:    []any{},
	}, nil
}

func (s *Service) completion(projectID uuid.UUID) (float64, TaskCounts, error) {
	if s.stats == nil {
		return 0, TaskCounts{}, nil
	}
	return s.stats.Completion(projectID)
}

func validateStatusPriority(status, priority string) error {
	switch status {
	case StatusDraft, StatusActive, StatusOnHold, StatusCompleted, StatusArchived:
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

func conflictOr(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return apperr.ErrConflict.WithMessage("user is already a project member")
	}
	if err != nil && strings.Contains(strings.ToLower(err.Error()), "unique") {
		return apperr.ErrConflict.WithMessage("user is already a project member")
	}
	return err
}
