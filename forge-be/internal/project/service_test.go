package project

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"workspace/internal/department"
	"workspace/internal/rbac/perm"
	"workspace/internal/user"
	"workspace/pkg/apperr"
	"workspace/pkg/authctx"
)

type fakeStore struct {
	projects map[uuid.UUID]*Project
	members  map[uuid.UUID][]Member
}

func newFakeStore() *fakeStore {
	return &fakeStore{projects: map[uuid.UUID]*Project{}, members: map[uuid.UUID][]Member{}}
}

func (f *fakeStore) Create(p *Project) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	cp := *p
	f.projects[p.ID] = &cp
	f.members[p.ID] = []Member{{ID: uuid.New(), ProjectID: p.ID, UserID: p.OwnerID, Role: RoleLead}}
	return nil
}
func (f *fakeStore) Save(p *Project) error {
	cp := *p
	f.projects[p.ID] = &cp
	return nil
}
func (f *fakeStore) GetByID(id uuid.UUID) (*Project, error) {
	p, ok := f.projects[id]
	if !ok {
		return nil, apperr.ErrNotFound.WithMessage("project not found")
	}
	cp := *p
	return &cp, nil
}
func (f *fakeStore) List(flt ListFilter) ([]Project, int64, error) {
	var out []Project
	for _, p := range f.projects {
		if p.DeletedAt.Valid && flt.Status != StatusArchived {
			continue
		}
		if flt.Status != "" && p.Status != flt.Status {
			continue
		}
		if !flt.ScopeAll && p.OwnerID != flt.UserID {
			ok := false
			for _, m := range f.members[p.ID] {
				if m.UserID == flt.UserID {
					ok = true
					break
				}
			}
			if !ok {
				continue
			}
		}
		out = append(out, *p)
	}
	return out, int64(len(out)), nil
}
func (f *fakeStore) Archive(id uuid.UUID) error {
	p, err := f.GetByID(id)
	if err != nil {
		return err
	}
	p.Status = StatusArchived
	p.DeletedAt.Valid = true
	f.projects[id] = p
	return nil
}
func (f *fakeStore) GetMember(projectID, userID uuid.UUID) (*Member, error) {
	for i := range f.members[projectID] {
		if f.members[projectID][i].UserID == userID {
			m := f.members[projectID][i]
			return &m, nil
		}
	}
	return nil, apperr.ErrNotFound.WithMessage("member not found")
}
func (f *fakeStore) ListMembers(projectID uuid.UUID, _, _ int) ([]Member, int64, error) {
	rows := f.members[projectID]
	return rows, int64(len(rows)), nil
}
func (f *fakeStore) CountMembers(projectID uuid.UUID) (int64, error) {
	return int64(len(f.members[projectID])), nil
}
func (f *fakeStore) AddMember(m *Member) error {
	if _, err := f.GetMember(m.ProjectID, m.UserID); err == nil {
		return apperr.ErrConflict.WithMessage("user is already a project member")
	}
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	f.members[m.ProjectID] = append(f.members[m.ProjectID], *m)
	return nil
}
func (f *fakeStore) SaveMember(m *Member) error { return nil }
func (f *fakeStore) DeleteMember(projectID, userID uuid.UUID) error {
	rows := f.members[projectID]
	out := rows[:0]
	found := false
	for _, m := range rows {
		if m.UserID == userID {
			found = true
			continue
		}
		out = append(out, m)
	}
	if !found {
		return apperr.ErrNotFound
	}
	f.members[projectID] = out
	return nil
}
func (f *fakeStore) IsMember(projectID, userID uuid.UUID) (bool, string, error) {
	m, err := f.GetMember(projectID, userID)
	if err != nil {
		return false, "", nil
	}
	return true, m.Role, nil
}

type fakeUsers struct{ ids map[uuid.UUID]bool }

func (f fakeUsers) GetByID(id uuid.UUID) (*user.User, error) {
	if f.ids[id] {
		return &user.User{ID: id, Name: "U"}, nil
	}
	return nil, apperr.ErrNotFound
}

type fakeDepts struct{}

func (fakeDepts) GetByID(uuid.UUID) (*department.Department, error) {
	return nil, apperr.ErrNotFound
}

type fakeStats struct{ pct float64 }

func (f fakeStats) Completion(uuid.UUID) (float64, TaskCounts, error) {
	return f.pct, TaskCounts{Done: 1, Total: 2}, nil
}

func TestProjectCreateAndRBAC(t *testing.T) {
	owner := uuid.New()
	other := uuid.New()
	store := newFakeStore()
	svc := NewService(store, fakeUsers{ids: map[uuid.UUID]bool{owner: true, other: true}}, fakeDepts{}, fakeStats{pct: 50}, nil)
	pm := authctx.Principal{UserID: owner, RoleCode: perm.RoleProjectManager, Permissions: []string{perm.ProjectCreate, perm.ProjectDelete}}
	member := authctx.Principal{UserID: other, RoleCode: perm.RoleMember, Permissions: []string{perm.TaskManage}}

	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	p, err := svc.Create(context.Background(), pm, "1.1.1.1", CreateInput{
		Name: "Alpha", OwnerID: owner, StartDate: start, TargetEndDate: end,
	})
	require.NoError(t, err)
	require.Equal(t, StatusDraft, p.Status)

	_, err = svc.Create(context.Background(), pm, "", CreateInput{
		Name: "Bad", OwnerID: owner, StartDate: end, TargetEndDate: start,
	})
	require.Equal(t, apperr.ErrValidation.Code, mustCode(err))

	_, _, err = svc.Get(member, p.ID)
	require.Equal(t, apperr.ErrForbidden.Code, mustCode(err))

	_, err = svc.AddMember(context.Background(), pm, "", p.ID, other, RoleMember)
	require.NoError(t, err)
	sum, err := func() (Summary, error) {
		_, s, err := svc.Get(member, p.ID)
		return s, err
	}()
	require.NoError(t, err)
	require.Equal(t, 50.0, sum.CompletionPercent)
	require.Empty(t, sum.Activity)

	err = svc.RemoveMember(context.Background(), pm, "", p.ID, owner)
	require.Equal(t, apperr.ErrValidation.Code, mustCode(err))

	_, err = svc.Patch(context.Background(), pm, "", p.ID, PatchInput{Status: strPtr(StatusArchived)})
	require.Equal(t, apperr.ErrValidation.Code, mustCode(err))

	active := StatusActive
	_, err = svc.Patch(context.Background(), pm, "", p.ID, PatchInput{Status: &active})
	require.NoError(t, err)
	hold := StatusOnHold
	_, err = svc.Patch(context.Background(), pm, "", p.ID, PatchInput{Status: &hold})
	require.NoError(t, err)
	require.NoError(t, svc.Delete(context.Background(), pm, "", p.ID))
}

func strPtr(s string) *string { return &s }

func mustCode(err error) string {
	ae, ok := apperr.As(err)
	if !ok {
		return ""
	}
	return ae.Code
}
