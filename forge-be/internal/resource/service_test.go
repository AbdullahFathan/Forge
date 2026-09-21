package resource

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"workspace/internal/project"
	"workspace/internal/rbac/perm"
	"workspace/internal/task"
	"workspace/internal/user"
	"workspace/pkg/apperr"
	"workspace/pkg/authctx"
)

type fakeStore struct {
	allocs   map[uuid.UUID]*Allocation
	holidays map[uuid.UUID]*Holiday
	users    map[uuid.UUID]*user.User
}

func newFakeStore() *fakeStore {
	return &fakeStore{
		allocs:   map[uuid.UUID]*Allocation{},
		holidays: map[uuid.UUID]*Holiday{},
		users:    map[uuid.UUID]*user.User{},
	}
}

func (f *fakeStore) Create(a *Allocation) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	cp := *a
	f.allocs[a.ID] = &cp
	return nil
}
func (f *fakeStore) Save(a *Allocation) error {
	cp := *a
	f.allocs[a.ID] = &cp
	return nil
}
func (f *fakeStore) GetByID(id uuid.UUID) (*Allocation, error) {
	a, ok := f.allocs[id]
	if !ok {
		return nil, apperr.ErrNotFound.WithMessage("allocation not found")
	}
	cp := *a
	return &cp, nil
}
func (f *fakeStore) SoftDelete(id uuid.UUID) error {
	if _, ok := f.allocs[id]; !ok {
		return apperr.ErrNotFound.WithMessage("allocation not found")
	}
	delete(f.allocs, id)
	return nil
}
func (f *fakeStore) List(flt ListFilter) ([]Allocation, int64, error) {
	var out []Allocation
	for _, a := range f.allocs {
		if flt.UserID != nil && a.UserID != *flt.UserID {
			continue
		}
		if flt.ProjectID != nil && a.ProjectID != *flt.ProjectID {
			continue
		}
		out = append(out, *a)
	}
	return out, int64(len(out)), nil
}
func (f *fakeStore) OverlappingSameProject(userID, projectID uuid.UUID, from, to time.Time, excludeID *uuid.UUID) ([]Allocation, error) {
	var out []Allocation
	for _, a := range f.allocs {
		if a.UserID != userID || a.ProjectID != projectID {
			continue
		}
		if excludeID != nil && a.ID == *excludeID {
			continue
		}
		if Overlaps(a.StartDate, a.EndDate, from, to) {
			out = append(out, *a)
		}
	}
	return out, nil
}
func (f *fakeStore) ForUsersInRange(ids []uuid.UUID, from, to time.Time) ([]Allocation, error) {
	set := map[uuid.UUID]struct{}{}
	for _, id := range ids {
		set[id] = struct{}{}
	}
	var out []Allocation
	for _, a := range f.allocs {
		if _, ok := set[a.UserID]; !ok {
			continue
		}
		if Overlaps(a.StartDate, a.EndDate, from, to) {
			out = append(out, *a)
		}
	}
	return out, nil
}
func (f *fakeStore) AllInRange(from, to time.Time) ([]Allocation, error) {
	var out []Allocation
	for _, a := range f.allocs {
		if Overlaps(a.StartDate, a.EndDate, from, to) {
			out = append(out, *a)
		}
	}
	return out, nil
}
func (f *fakeStore) ListHolidays() ([]Holiday, error) {
	var out []Holiday
	for _, h := range f.holidays {
		out = append(out, *h)
	}
	return out, nil
}
func (f *fakeStore) CreateHoliday(h *Holiday) error {
	if h.ID == uuid.Nil {
		h.ID = uuid.New()
	}
	cp := *h
	f.holidays[h.ID] = &cp
	return nil
}
func (f *fakeStore) GetHoliday(id uuid.UUID) (*Holiday, error) {
	h, ok := f.holidays[id]
	if !ok {
		return nil, apperr.ErrNotFound.WithMessage("holiday not found")
	}
	cp := *h
	return &cp, nil
}
func (f *fakeStore) SaveHoliday(h *Holiday) error {
	cp := *h
	f.holidays[h.ID] = &cp
	return nil
}
func (f *fakeStore) GetHolidayByDate(dt time.Time) (*Holiday, error) {
	for _, h := range f.holidays {
		if DateUTC(h.Date).Equal(DateUTC(dt)) {
			cp := *h
			return &cp, nil
		}
	}
	return nil, apperr.ErrNotFound.WithMessage("holiday not found")
}
func (f *fakeStore) DeleteHoliday(id uuid.UUID) error {
	if _, ok := f.holidays[id]; !ok {
		return apperr.ErrNotFound.WithMessage("holiday not found")
	}
	delete(f.holidays, id)
	return nil
}
func (f *fakeStore) ListActiveUsers(dept *uuid.UUID, skill string) ([]user.User, error) {
	var out []user.User
	for _, u := range f.users {
		if !u.IsActive {
			continue
		}
		if dept != nil && (u.DepartmentID == nil || *u.DepartmentID != *dept) {
			continue
		}
		if skill != "" {
			ok := false
			for _, s := range u.Skills {
				if s.Skill == skill {
					ok = true
					break
				}
			}
			if !ok {
				continue
			}
		}
		out = append(out, *u)
	}
	return out, nil
}
func (f *fakeStore) GetUser(id uuid.UUID) (*user.User, error) {
	u, ok := f.users[id]
	if !ok {
		return nil, apperr.ErrNotFound.WithMessage("user not found")
	}
	cp := *u
	return &cp, nil
}

type fakeProjects struct {
	projects map[uuid.UUID]*project.Project
	leads    map[uuid.UUID]uuid.UUID
	upserts  int
}

func (f *fakeProjects) MustSee(actor authctx.Principal, projectID uuid.UUID) (*project.Project, string, error) {
	p, ok := f.projects[projectID]
	if !ok {
		return nil, "", apperr.ErrNotFound
	}
	if authctx.HasPermission(actor, perm.ProjectReadAll) || p.OwnerID == actor.UserID {
		return p, project.RoleLead, nil
	}
	return nil, "", apperr.ErrForbidden
}
func (f *fakeProjects) MustManage(actor authctx.Principal, projectID uuid.UUID) (*project.Project, error) {
	p, _, err := f.MustSee(actor, projectID)
	if err != nil {
		return nil, err
	}
	if actor.RoleCode == perm.RoleSuperAdmin || actor.RoleCode == perm.RoleAdmin || p.OwnerID == actor.UserID {
		return p, nil
	}
	return nil, apperr.ErrForbidden
}
func (f *fakeProjects) UpsertMemberRole(uuid.UUID, uuid.UUID, string) error {
	f.upserts++
	return nil
}

type fakeTasks struct{}

func (fakeTasks) ListByAssignee(uuid.UUID) ([]task.Task, error) { return nil, nil }

type fixedClock struct{ t time.Time }

func (f fixedClock) Now() time.Time { return f.t }

func TestCreateOverlapRejectedAndOverAllocationWarned(t *testing.T) {
	store := newFakeStore()
	uid, owner, projA, projB := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	store.users[uid] = &user.User{ID: uid, Name: "Dev", IsActive: true, CapacityHoursPerDay: 8}
	projs := &fakeProjects{projects: map[uuid.UUID]*project.Project{
		projA: {ID: projA, OwnerID: owner, Name: "A"},
		projB: {ID: projB, OwnerID: owner, Name: "B"},
	}}
	svc := NewService(store, projs, fakeTasks{}, nil, nil)
	pm := authctx.Principal{UserID: owner, RoleCode: perm.RoleProjectManager, Permissions: []string{perm.ResourceAllocate, perm.ProjectCreate}}

	from, to := d("2026-03-02"), d("2026-03-20")
	_, _, _, err := svc.Create(context.Background(), pm, "", CreateInput{
		UserID: uid, ProjectID: projA, AllocationPercent: 50, StartDate: from, EndDate: to, Role: RoleMember,
	})
	require.NoError(t, err)
	_, warn, over, err := svc.Create(context.Background(), pm, "", CreateInput{
		UserID: uid, ProjectID: projB, AllocationPercent: 60, StartDate: from, EndDate: to, Role: RoleMember,
	})
	require.NoError(t, err)
	require.True(t, over)
	require.Equal(t, []string{WarningOverAllocated}, warn)
	require.Equal(t, 2, projs.upserts)

	_, _, _, err = svc.Create(context.Background(), pm, "", CreateInput{
		UserID: uid, ProjectID: projA, AllocationPercent: 10, StartDate: from, EndDate: to, Role: RoleMember,
	})
	require.Equal(t, apperr.ErrConflict.Code, mustCode(err))
}

func TestPMCannotAllocateOthersProject(t *testing.T) {
	store := newFakeStore()
	uid, owner, other, projID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	store.users[uid] = &user.User{ID: uid, Name: "Dev", IsActive: true, CapacityHoursPerDay: 8}
	projs := &fakeProjects{projects: map[uuid.UUID]*project.Project{
		projID: {ID: projID, OwnerID: owner, Name: "A"},
	}}
	svc := NewService(store, projs, fakeTasks{}, nil, nil)
	pm := authctx.Principal{UserID: other, RoleCode: perm.RoleProjectManager, Permissions: []string{perm.ResourceAllocate}}
	_, _, _, err := svc.Create(context.Background(), pm, "", CreateInput{
		UserID: uid, ProjectID: projID, AllocationPercent: 50, StartDate: d("2026-03-02"), EndDate: d("2026-03-20"),
	})
	require.Equal(t, apperr.ErrForbidden.Code, mustCode(err))
}

func TestForecastRangeAndOverloadClock(t *testing.T) {
	store := newFakeStore()
	uid := uuid.New()
	projID := uuid.New()
	store.users[uid] = &user.User{ID: uid, Name: "Dev", IsActive: true, CapacityHoursPerDay: 8}
	store.allocs[uuid.New()] = &Allocation{
		ID: uuid.New(), UserID: uid, ProjectID: projID, AllocationPercent: 110,
		StartDate: d("2026-03-14"), EndDate: d("2026-03-14"), Role: RoleMember,
	}
	svc := NewService(store, &fakeProjects{projects: map[uuid.UUID]*project.Project{}}, fakeTasks{}, nil, fixedClock{t: d("2026-03-01")})
	_, err := svc.Forecast(RangeQuery{From: d("2026-03-02"), To: d("2026-03-10"), Granularity: "week"})
	require.Equal(t, apperr.ErrValidation.Code, mustCode(err))

	alerts, err := svc.OverloadAlerts()
	require.NoError(t, err)
	require.Len(t, alerts, 1)

	svc15 := NewService(store, &fakeProjects{projects: map[uuid.UUID]*project.Project{}}, fakeTasks{}, nil, fixedClock{t: d("2026-02-27")})
	alerts, err = svc15.OverloadAlerts()
	require.NoError(t, err)
	require.Empty(t, alerts) // 2026-03-14 is day 15 from 2026-02-27 (Feb 27 + 13 = Mar 12)
}

func TestWorkloadSelfOnly(t *testing.T) {
	store := newFakeStore()
	uid, other := uuid.New(), uuid.New()
	store.users[uid] = &user.User{ID: uid, Name: "Me", IsActive: true, CapacityHoursPerDay: 8}
	store.users[other] = &user.User{ID: other, Name: "Them", IsActive: true, CapacityHoursPerDay: 8}
	svc := NewService(store, &fakeProjects{projects: map[uuid.UUID]*project.Project{}}, fakeTasks{}, nil, fixedClock{t: d("2026-03-02")})
	member := authctx.Principal{UserID: uid, RoleCode: perm.RoleMember, Permissions: []string{perm.TaskManage}}
	_, err := svc.Workload(member, &other)
	require.Equal(t, apperr.ErrForbidden.Code, mustCode(err))
	wl, err := svc.Workload(member, nil)
	require.NoError(t, err)
	require.Equal(t, uid, wl.UserID)
	require.Len(t, wl.WeekSeries, 4)
}

func TestPatchHolidayConflict(t *testing.T) {
	store := newFakeStore()
	svc := NewService(store, &fakeProjects{}, fakeTasks{}, nil, nil)
	a, err := svc.CreateHoliday(d("2026-05-01"), "Labor")
	require.NoError(t, err)
	_, err = svc.CreateHoliday(d("2026-08-17"), "Independence")
	require.NoError(t, err)
	next := d("2026-08-17")
	_, err = svc.PatchHoliday(a.ID, &next, nil)
	require.Equal(t, apperr.ErrConflict.Code, mustCode(err))
	name := "May Day"
	h, err := svc.PatchHoliday(a.ID, nil, &name)
	require.NoError(t, err)
	require.Equal(t, "May Day", h.Name)
}

func mustCode(err error) string {
	ae, ok := apperr.As(err)
	if !ok {
		return ""
	}
	return ae.Code
}
