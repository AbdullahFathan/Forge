package task

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"workspace/internal/project"
	"workspace/internal/rbac/perm"
	"workspace/pkg/apperr"
	"workspace/pkg/authctx"
)

type fakeTasks struct {
	byID      map[uuid.UUID]*Task
	children  map[uuid.UUID]int
	deps      []Dependency
	assignees map[uuid.UUID][]uuid.UUID
}

func newFakeTasks() *fakeTasks {
	return &fakeTasks{byID: map[uuid.UUID]*Task{}, children: map[uuid.UUID]int{}, assignees: map[uuid.UUID][]uuid.UUID{}}
}

func (f *fakeTasks) Create(t *Task, ids []uuid.UUID) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	cp := *t
	f.byID[t.ID] = &cp
	f.assignees[t.ID] = append([]uuid.UUID{}, ids...)
	if t.ParentTaskID != nil {
		f.children[*t.ParentTaskID]++
	}
	return nil
}
func (f *fakeTasks) Save(t *Task, ids *[]uuid.UUID) error {
	cp := *t
	f.byID[t.ID] = &cp
	if ids != nil {
		f.assignees[t.ID] = append([]uuid.UUID{}, *ids...)
	}
	return nil
}
func (f *fakeTasks) GetByID(id uuid.UUID, _ bool) (*Task, error) {
	t, ok := f.byID[id]
	if !ok {
		return nil, apperr.ErrNotFound
	}
	cp := *t
	return &cp, nil
}
func (f *fakeTasks) List(ListFilter) ([]Task, int64, error) { return nil, 0, nil }
func (f *fakeTasks) SoftDelete(id uuid.UUID) error {
	delete(f.byID, id)
	return nil
}
func (f *fakeTasks) HasChildren(id uuid.UUID) (bool, error)     { return f.children[id] > 0, nil }
func (f *fakeTasks) MaxPosition(uuid.UUID, string) (int, error) { return 0, nil }
func (f *fakeTasks) AddDependency(d *Dependency) error {
	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}
	f.deps = append(f.deps, *d)
	return nil
}
func (f *fakeTasks) DeleteDependency(uuid.UUID, uuid.UUID) error { return nil }
func (f *fakeTasks) DependencyEdges(uuid.UUID) (map[uuid.UUID][]uuid.UUID, error) {
	m := map[uuid.UUID][]uuid.UUID{}
	for _, d := range f.deps {
		m[d.TaskID] = append(m[d.TaskID], d.DependsOnTaskID)
	}
	return m, nil
}
func (f *fakeTasks) UnfinishedPredecessors(taskID uuid.UUID) (bool, error) {
	for _, d := range f.deps {
		if d.TaskID != taskID {
			continue
		}
		pred := f.byID[d.DependsOnTaskID]
		if pred != nil && pred.Status != StatusDone {
			return true, nil
		}
	}
	return false, nil
}
func (f *fakeTasks) AddComment(*Comment) error                 { return nil }
func (f *fakeTasks) ListComments(uuid.UUID) ([]Comment, error) { return nil, nil }
func (f *fakeTasks) IsAssignee(taskID, userID uuid.UUID) (bool, error) {
	for _, id := range f.assignees[taskID] {
		if id == userID {
			return true, nil
		}
	}
	return false, nil
}

type fakeProj struct {
	projectID uuid.UUID
	members   map[uuid.UUID]string
	manage    map[uuid.UUID]bool
}

func (f fakeProj) MustSee(actor authctx.Principal, projectID uuid.UUID) (*project.Project, string, error) {
	if projectID != f.projectID {
		return nil, "", apperr.ErrNotFound
	}
	role := f.members[actor.UserID]
	if role == "" && actor.UserID != (uuid.UUID{}) {
		if _, ok := f.members[actor.UserID]; !ok && !authctx.HasPermission(actor, perm.ProjectReadAll) {
			if f.manage[actor.UserID] {
				return &project.Project{ID: projectID, OwnerID: actor.UserID}, project.RoleLead, nil
			}
		}
	}
	if role == "" && !f.manage[actor.UserID] && !authctx.HasPermission(actor, perm.ProjectReadAll) {
		return nil, "", apperr.ErrForbidden
	}
	if role == "" {
		role = project.RoleLead
	}
	return &project.Project{ID: projectID}, role, nil
}
func (f fakeProj) MustManage(actor authctx.Principal, projectID uuid.UUID) (*project.Project, error) {
	if !f.manage[actor.UserID] {
		return nil, apperr.ErrForbidden
	}
	return &project.Project{ID: projectID, OwnerID: actor.UserID}, nil
}
func (f fakeProj) MemberRole(projectID, userID uuid.UUID) (bool, string, error) {
	role, ok := f.members[userID]
	return ok, role, nil
}

func TestSubtaskDepthAndFSGuard(t *testing.T) {
	pid := uuid.New()
	pmID := uuid.New()
	memberID := uuid.New()
	store := newFakeTasks()
	proj := fakeProj{
		projectID: pid,
		members:   map[uuid.UUID]string{pmID: project.RoleLead, memberID: project.RoleMember},
		manage:    map[uuid.UUID]bool{pmID: true},
	}
	svc := NewService(store, proj, nil)
	pm := authctx.Principal{UserID: pmID, RoleCode: perm.RoleProjectManager, Permissions: []string{perm.TaskManage}}

	parent, err := svc.Create(context.Background(), pm, "", CreateInput{ProjectID: pid, Name: "P"})
	require.NoError(t, err)
	child, err := svc.Create(context.Background(), pm, "", CreateInput{ProjectID: pid, Name: "C", ParentTaskID: &parent.ID})
	require.NoError(t, err)
	_, err = svc.Create(context.Background(), pm, "", CreateInput{ProjectID: pid, Name: "G", ParentTaskID: &child.ID})
	require.Equal(t, apperr.ErrValidation.Code, mustCode(err))

	pred, err := svc.Create(context.Background(), pm, "", CreateInput{ProjectID: pid, Name: "Pred"})
	require.NoError(t, err)
	dep, err := svc.Create(context.Background(), pm, "", CreateInput{ProjectID: pid, Name: "Dep"})
	require.NoError(t, err)
	_, err = svc.AddDependency(context.Background(), pm, "", dep.ID, pred.ID)
	require.NoError(t, err)
	_, err = svc.AddDependency(context.Background(), pm, "", pred.ID, dep.ID)
	require.Equal(t, apperr.ErrValidation.Code, mustCode(err))

	st := StatusInProgress
	_, err = svc.Patch(context.Background(), pm, "", dep.ID, PatchInput{Status: &st})
	require.Equal(t, apperr.ErrValidation.Code, mustCode(err))

	done := StatusDone
	todo := StatusTodo
	_, err = svc.Patch(context.Background(), pm, "", pred.ID, PatchInput{Status: &todo})
	require.NoError(t, err)
	_, err = svc.Patch(context.Background(), pm, "", pred.ID, PatchInput{Status: &st})
	require.NoError(t, err)
	_, err = svc.Patch(context.Background(), pm, "", pred.ID, PatchInput{Status: &done})
	require.NoError(t, err)
	_, err = svc.Patch(context.Background(), pm, "", dep.ID, PatchInput{Status: &todo})
	require.NoError(t, err)
}

func mustCode(err error) string {
	ae, ok := apperr.As(err)
	if !ok {
		return ""
	}
	return ae.Code
}
