package rbac

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"workspace/internal/rbac/perm"
	"workspace/pkg/apperr"
)

type memStore struct {
	roles      map[uuid.UUID]*Role
	byCode     map[string]uuid.UUID
	perms      map[string]Permission
	userCounts map[uuid.UUID]int64
}

func newMem() *memStore {
	m := &memStore{
		roles:      map[uuid.UUID]*Role{},
		byCode:     map[string]uuid.UUID{},
		perms:      map[string]Permission{},
		userCounts: map[uuid.UUID]int64{},
	}
	for _, c := range perm.All() {
		m.perms[c] = Permission{ID: uuid.New(), Code: c, Name: c}
	}
	sys := &Role{ID: uuid.New(), Code: perm.RoleAdmin, Name: "Admin", IsSystem: true}
	m.roles[sys.ID] = sys
	m.byCode[sys.Code] = sys.ID
	return m
}

func (m *memStore) clone(r *Role) *Role {
	cp := *r
	cp.Permissions = append([]Permission{}, r.Permissions...)
	return &cp
}

func (m *memStore) ListRoles() ([]Role, error) {
	out := make([]Role, 0, len(m.roles))
	for _, r := range m.roles {
		out = append(out, *m.clone(r))
	}
	return out, nil
}
func (m *memStore) GetRoleByID(id uuid.UUID) (*Role, error) {
	r, ok := m.roles[id]
	if !ok {
		return nil, apperr.ErrNotFound.WithMessage("role not found")
	}
	return m.clone(r), nil
}
func (m *memStore) GetRoleByCode(code string) (*Role, error) {
	id, ok := m.byCode[code]
	if !ok {
		return nil, apperr.ErrNotFound.WithMessage("role not found")
	}
	return m.GetRoleByID(id)
}
func (m *memStore) CreateRole(role *Role) error {
	if role.ID == uuid.Nil {
		role.ID = uuid.New()
	}
	m.roles[role.ID] = m.clone(role)
	m.byCode[role.Code] = role.ID
	return nil
}
func (m *memStore) SaveRole(role *Role) error {
	m.roles[role.ID] = m.clone(role)
	return nil
}
func (m *memStore) DeleteRole(id uuid.UUID) error {
	r, ok := m.roles[id]
	if !ok {
		return apperr.ErrNotFound.WithMessage("role not found")
	}
	delete(m.byCode, r.Code)
	delete(m.roles, id)
	return nil
}
func (m *memStore) ListPermissions() ([]Permission, error) {
	out := make([]Permission, 0, len(m.perms))
	for _, p := range m.perms {
		out = append(out, p)
	}
	return out, nil
}
func (m *memStore) PermissionsByCodes(codes []string) ([]Permission, error) {
	var out []Permission
	for _, c := range codes {
		if p, ok := m.perms[c]; ok {
			out = append(out, p)
		}
	}
	return out, nil
}
func (m *memStore) ReplacePermissions(role *Role, perms []Permission) error {
	r := m.roles[role.ID]
	r.Permissions = append([]Permission{}, perms...)
	return nil
}
func (m *memStore) CountUsersByRole(id uuid.UUID) (int64, error) {
	return m.userCounts[id], nil
}

func TestCreateRoleRejectsInvalidPermission(t *testing.T) {
	svc := NewService(newMem())
	_, err := svc.Create(CreateInput{Name: "Ops", PermissionCodes: []string{"not.a.perm"}})
	require.Error(t, err)
	ae, ok := apperr.As(err)
	require.True(t, ok)
	require.Equal(t, apperr.ErrValidation.Code, ae.Code)
}

func TestCreateAndAssignCustomRole(t *testing.T) {
	svc := NewService(newMem())
	role, err := svc.Create(CreateInput{Name: "Ops Lead", PermissionCodes: []string{perm.TaskManage, perm.ReportExport}})
	require.NoError(t, err)
	require.False(t, role.IsSystem)
	require.Equal(t, "OPS_LEAD", role.Code)
	require.Len(t, role.Permissions, 2)
}

func TestCannotMutateOrDeleteSystemRole(t *testing.T) {
	store := newMem()
	svc := NewService(store)
	var sysID uuid.UUID
	for id, r := range store.roles {
		if r.IsSystem {
			sysID = id
			break
		}
	}
	name := "Nope"
	_, err := svc.Patch(sysID, PatchInput{Name: &name})
	require.Error(t, err)
	ae, _ := apperr.As(err)
	require.Equal(t, apperr.ErrForbidden.Code, ae.Code)
	require.Error(t, svc.Delete(sysID))
}

func TestDeleteRoleAssignedUsersConflict(t *testing.T) {
	store := newMem()
	svc := NewService(store)
	role, err := svc.Create(CreateInput{Name: "Temp", PermissionCodes: []string{perm.TaskManage}})
	require.NoError(t, err)
	store.userCounts[role.ID] = 1
	err = svc.Delete(role.ID)
	require.Error(t, err)
	ae, _ := apperr.As(err)
	require.Equal(t, apperr.ErrConflict.Code, ae.Code)
}

func TestCannotReuseSystemCode(t *testing.T) {
	svc := NewService(newMem())
	_, err := svc.Create(CreateInput{Name: "X", Code: perm.RoleAdmin, PermissionCodes: []string{perm.TaskManage}})
	require.Error(t, err)
}
