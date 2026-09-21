package jobs

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"workspace/internal/notification"
	"workspace/internal/project"
	"workspace/internal/rbac/perm"
	"workspace/internal/resource"
	"workspace/internal/task"
)

type Clock interface {
	Now() time.Time
}

type realClock struct{}

func (realClock) Now() time.Time { return time.Now().UTC() }

type Emitter interface {
	Emit(ctx context.Context, ev notification.Event) error
}

type Tasks interface {
	ListOpenDueBetween(from, to time.Time) ([]task.Task, error)
	ListOpenOverdue(before time.Time) ([]task.Task, error)
}

type Projects interface {
	DueBetween(from, to time.Time) ([]project.Project, error)
	GetByID(uuid.UUID) (*project.Project, error)
}

type Resources interface {
	OverloadAlerts() ([]resource.OverloadAlert, error)
	AllocationsForUser(userID uuid.UUID, from, to time.Time) ([]resource.Allocation, error)
}

type PermUsers interface {
	ListUserIDsWithPermission(code string) ([]uuid.UUID, error)
}

type Scheduler struct {
	clock Clock
	rdb   *redis.Client
	emit  Emitter
	tasks Tasks
	projs Projects
	res   Resources
	perms PermUsers
}

func New(clock Clock, rdb *redis.Client, emit Emitter, tasks Tasks, projs Projects, res Resources, perms PermUsers) *Scheduler {
	if clock == nil {
		clock = realClock{}
	}
	return &Scheduler{clock: clock, rdb: rdb, emit: emit, tasks: tasks, projs: projs, res: res, perms: perms}
}

func (s *Scheduler) Run(ctx context.Context, interval time.Duration) {
	if interval <= 0 {
		interval = 15 * time.Minute
	}
	t := time.NewTicker(interval)
	defer t.Stop()
	_ = s.RunOnce(ctx)
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			_ = s.RunOnce(ctx)
		}
	}
}

func (s *Scheduler) RunOnce(ctx context.Context) error {
	if s.rdb != nil {
		ok, err := s.rdb.SetNX(ctx, "jobs:notify", "1", 10*time.Minute).Result()
		if err != nil || !ok {
			return err
		}
	}
	now := s.clock.Now().UTC()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	if err := s.dueSoon(ctx, today); err != nil {
		return err
	}
	if err := s.overdue(ctx, today); err != nil {
		return err
	}
	if err := s.projectDue(ctx, today); err != nil {
		return err
	}
	return s.overload(ctx, today)
}

func (s *Scheduler) dueSoon(ctx context.Context, today time.Time) error {
	rows, err := s.tasks.ListOpenDueBetween(today, today.AddDate(0, 0, 3))
	if err != nil {
		return err
	}
	for _, t := range rows {
		p, _ := s.projs.GetByID(t.ProjectID)
		for _, u := range t.Assignees {
			_ = s.emit.Emit(ctx, notification.Event{
				UserID: u.ID, Type: notification.TypeTaskDueSoon,
				Title: "Task due soon", Body: t.Name, EntityID: t.ID, Day: today,
				Metadata: map[string]any{"taskId": t.ID.String()},
			})
		}
		if p != nil {
			_ = s.emit.Emit(ctx, notification.Event{
				UserID: p.OwnerID, Type: notification.TypeTaskDueSoon,
				Title: "Task due soon", Body: t.Name, EntityID: t.ID, Day: today,
				Metadata: map[string]any{"taskId": t.ID.String(), "projectId": t.ProjectID.String()},
			})
		}
	}
	return nil
}

func (s *Scheduler) overdue(ctx context.Context, today time.Time) error {
	rows, err := s.tasks.ListOpenOverdue(today)
	if err != nil {
		return err
	}
	for _, t := range rows {
		p, _ := s.projs.GetByID(t.ProjectID)
		for _, u := range t.Assignees {
			_ = s.emit.Emit(ctx, notification.Event{
				UserID: u.ID, Type: notification.TypeTaskOverdue,
				Title: "Task overdue", Body: t.Name, EntityID: t.ID, Day: today,
			})
		}
		if p != nil {
			_ = s.emit.Emit(ctx, notification.Event{
				UserID: p.OwnerID, Type: notification.TypeTaskOverdue,
				Title: "Task overdue", Body: t.Name, EntityID: t.ID, Day: today,
			})
		}
	}
	return nil
}

func (s *Scheduler) projectDue(ctx context.Context, today time.Time) error {
	from, to := today.AddDate(0, 0, 1), today.AddDate(0, 0, 14)
	rows, err := s.projs.DueBetween(from, to)
	if err != nil {
		return err
	}
	for _, p := range rows {
		_ = s.emit.Emit(ctx, notification.Event{
			UserID: p.OwnerID, Type: notification.TypeProjectDueSoon,
			Title: "Project approaching deadline", Body: p.Name, EntityID: p.ID, Day: today,
			Metadata: map[string]any{"projectId": p.ID.String()},
		})
	}
	return nil
}

func (s *Scheduler) overload(ctx context.Context, today time.Time) error {
	alerts, err := s.res.OverloadAlerts()
	if err != nil {
		return err
	}
	rms, err := s.perms.ListUserIDsWithPermission(perm.CapacityView)
	if err != nil {
		return err
	}
	to := today.AddDate(0, 0, 13)
	for _, a := range alerts {
		for _, rm := range rms {
			_ = s.emit.Emit(ctx, notification.Event{
				UserID: rm, Type: notification.TypeResourceOverload,
				Title: "Resource over-allocated", Body: a.Name, EntityID: a.UserID, Day: today,
				Metadata: map[string]any{"userId": a.UserID.String(), "maxPercent": a.MaxPercent},
			})
		}
		allocs, err := s.res.AllocationsForUser(a.UserID, today, to)
		if err != nil {
			return err
		}
		seen := map[uuid.UUID]struct{}{}
		for _, al := range allocs {
			if al.Project == nil {
				p, err := s.projs.GetByID(al.ProjectID)
				if err != nil {
					continue
				}
				al.Project = p
			}
			oid := al.Project.OwnerID
			if _, ok := seen[oid]; ok {
				continue
			}
			seen[oid] = struct{}{}
			_ = s.emit.Emit(ctx, notification.Event{
				UserID: oid, Type: notification.TypeResourceOverload,
				Title: "Resource over-allocated", Body: a.Name, EntityID: a.UserID, Day: today,
			})
		}
	}
	return nil
}
