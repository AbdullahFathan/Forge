package rbac

import (
	"regexp"
	"strings"
	"unicode"

	"github.com/google/uuid"

	"workspace/internal/rbac/perm"
	"workspace/pkg/apperr"
)

type Store interface {
	ListRoles() ([]Role, error)
	GetRoleByID(uuid.UUID) (*Role, error)
	GetRoleByCode(string) (*Role, error)
	CreateRole(*Role) error
	SaveRole(*Role) error
	DeleteRole(uuid.UUID) error
	ListPermissions() ([]Permission, error)
	PermissionsByCodes([]string) ([]Permission, error)
	ReplacePermissions(*Role, []Permission) error
	CountUsersByRole(uuid.UUID) (int64, error)
}

type Service struct {
	store Store
}

func NewService(store Store) *Service {
	return &Service{store: store}
}

var codeRe = regexp.MustCompile(`^[A-Z][A-Z0-9_]{1,62}$`)

func (s *Service) List() ([]Role, error) {
	return s.store.ListRoles()
}

func (s *Service) Get(id uuid.UUID) (*Role, error) {
	return s.store.GetRoleByID(id)
}

func (s *Service) ListPermissions() ([]Permission, error) {
	return s.store.ListPermissions()
}

type CreateInput struct {
	Name            string
	Code            string
	PermissionCodes []string
}

func (s *Service) Create(in CreateInput) (*Role, error) {
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return nil, apperr.ErrValidation.WithMessage("name is required")
	}
	code, err := normalizeRoleCode(name, in.Code)
	if err != nil {
		return nil, err
	}
	if isSystemCode(code) {
		return nil, apperr.ErrValidation.WithMessage("cannot reuse a system role code")
	}
	if _, err := s.store.GetRoleByCode(code); err == nil {
		return nil, apperr.ErrConflict.WithMessage("role code already exists")
	} else if ae, ok := apperr.As(err); !ok || ae.Code != apperr.ErrNotFound.Code {
		return nil, err
	}
	perms, err := s.resolvePerms(in.PermissionCodes)
	if err != nil {
		return nil, err
	}
	role := &Role{Code: code, Name: name, IsSystem: false}
	if err := s.store.CreateRole(role); err != nil {
		return nil, err
	}
	if err := s.store.ReplacePermissions(role, perms); err != nil {
		return nil, err
	}
	return s.store.GetRoleByID(role.ID)
}

type PatchInput struct {
	Name            *string
	PermissionCodes *[]string
}

func (s *Service) Patch(id uuid.UUID, in PatchInput) (*Role, error) {
	role, err := s.store.GetRoleByID(id)
	if err != nil {
		return nil, err
	}
	if role.IsSystem {
		return nil, apperr.ErrForbidden.WithMessage("system roles cannot be modified")
	}
	if in.Name != nil {
		name := strings.TrimSpace(*in.Name)
		if name == "" {
			return nil, apperr.ErrValidation.WithMessage("name is required")
		}
		role.Name = name
		if err := s.store.SaveRole(role); err != nil {
			return nil, err
		}
	}
	if in.PermissionCodes != nil {
		perms, err := s.resolvePerms(*in.PermissionCodes)
		if err != nil {
			return nil, err
		}
		if err := s.store.ReplacePermissions(role, perms); err != nil {
			return nil, err
		}
	}
	return s.store.GetRoleByID(id)
}

func (s *Service) Delete(id uuid.UUID) error {
	role, err := s.store.GetRoleByID(id)
	if err != nil {
		return err
	}
	if role.IsSystem {
		return apperr.ErrForbidden.WithMessage("system roles cannot be deleted")
	}
	n, err := s.store.CountUsersByRole(id)
	if err != nil {
		return err
	}
	if n > 0 {
		return apperr.ErrConflict.WithMessage("role is still assigned to users")
	}
	return s.store.DeleteRole(id)
}

func (s *Service) resolvePerms(codes []string) ([]Permission, error) {
	seen := map[string]struct{}{}
	uniq := make([]string, 0, len(codes))
	for _, c := range codes {
		c = strings.TrimSpace(c)
		if c == "" {
			continue
		}
		if _, ok := seen[c]; ok {
			continue
		}
		seen[c] = struct{}{}
		uniq = append(uniq, c)
	}
	if len(uniq) == 0 {
		return nil, apperr.ErrValidation.WithMessage("at least one permission code is required")
	}
	rows, err := s.store.PermissionsByCodes(uniq)
	if err != nil {
		return nil, err
	}
	found := map[string]struct{}{}
	for _, p := range rows {
		found[p.Code] = struct{}{}
	}
	for _, c := range uniq {
		if _, ok := found[c]; !ok {
			return nil, apperr.ErrValidation.WithMessage("invalid permission code: " + c)
		}
	}
	return rows, nil
}

func normalizeRoleCode(name, code string) (string, error) {
	code = strings.ToUpper(strings.TrimSpace(code))
	if code == "" {
		code = slugCode(name)
	}
	if !codeRe.MatchString(code) {
		return "", apperr.ErrValidation.WithMessage("code must be uppercase letters, digits, and underscores")
	}
	return code, nil
}

func slugCode(name string) string {
	var b strings.Builder
	lastUS := true
	for _, r := range strings.ToUpper(name) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
			lastUS = false
			continue
		}
		if !lastUS {
			b.WriteByte('_')
			lastUS = true
		}
	}
	out := strings.Trim(b.String(), "_")
	if out == "" {
		return "CUSTOM"
	}
	if len(out) > 64 {
		out = out[:64]
	}
	return out
}

func isSystemCode(code string) bool {
	switch code {
	case perm.RoleSuperAdmin, perm.RoleAdmin, perm.RoleResourceManager, perm.RoleProjectManager, perm.RoleMember:
		return true
	default:
		return false
	}
}

func PermissionCodes(role *Role) []string {
	out := make([]string, 0, len(role.Permissions))
	for _, p := range role.Permissions {
		out = append(out, p.Code)
	}
	return out
}
