package jobs

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"workspace/internal/notification"
	"workspace/internal/project"
	"workspace/internal/resource"
	"workspace/internal/task"
	"workspace/internal/user"
)

type fixedClock struct{ t time.Time }

func (f fixedClock) Now() time.Time { return f.t }

type fakeTasks struct{ due, overdue []task.Task }

func (f fakeTasks) ListOpenDueBetween(_, _ time.Time) ([]task.Task, error) { return f.due, nil }
func (f fakeTasks) ListOpenOverdue(time.Time) ([]task.Task, error)         { return f.overdue, nil }

type fakeProjs struct{ p *project.Project }

func (f fakeProjs) DueBetween(_, _ time.Time) ([]project.Project, error) {
	if f.p == nil {
		return nil, nil
	}
	return []project.Project{*f.p}, nil
}
func (f fakeProjs) GetByID(uuid.UUID) (*project.Project, error) { return f.p, nil }

type fakeRes struct{}

func (fakeRes) OverloadAlerts() ([]resource.OverloadAlert, error) { return nil, nil }
func (fakeRes) AllocationsForUser(uuid.UUID, time.Time, time.Time) ([]resource.Allocation, error) {
	return nil, nil
}

type fakePerms struct{}

func (fakePerms) ListUserIDsWithPermission(string) ([]uuid.UUID, error) { return nil, nil }

func TestRunOnceDueSoon(t *testing.T) {
	today := time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC)
	assignee := uuid.New()
	owner := uuid.New()
	tid := uuid.New()
	pid := uuid.New()
	cap := &notification.Capture{}
	p := &project.Project{ID: pid, Name: "P", OwnerID: owner}
	s := New(fixedClock{t: today}, nil, cap, fakeTasks{due: []task.Task{{
		ID: tid, ProjectID: pid, Name: "T", Assignees: []user.User{{ID: assignee}},
	}}}, fakeProjs{p: p}, fakeRes{}, fakePerms{})
	require.NoError(t, s.RunOnce(context.Background()))
	var types []string
	for _, e := range cap.Events {
		if e.Type == notification.TypeTaskDueSoon {
			types = append(types, e.Type)
		}
	}
	require.GreaterOrEqual(t, len(types), 2)
}
