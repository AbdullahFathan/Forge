package rbac

import (
	"github.com/google/uuid"
	"gorm.io/gorm"

	"workspace/internal/rbac/perm"
	"workspace/pkg/apperr"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) ListRoles() ([]Role, error) {
	var rows []Role
	err := r.db.Preload("Permissions").Order("is_system DESC, name ASC").Find(&rows).Error
	return rows, err
}

func (r *Repository) CreateRole(role *Role) error {
	return r.db.Create(role).Error
}

func (r *Repository) SaveRole(role *Role) error {
	return r.db.Save(role).Error
}

func (r *Repository) DeleteRole(id uuid.UUID) error {
	res := r.db.Delete(&Role{}, "id = ?", id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return apperr.ErrNotFound.WithMessage("role not found")
	}
	return nil
}

func (r *Repository) ListPermissions() ([]Permission, error) {
	var rows []Permission
	err := r.db.Order("code ASC").Find(&rows).Error
	return rows, err
}

func (r *Repository) PermissionsByCodes(codes []string) ([]Permission, error) {
	if len(codes) == 0 {
		return []Permission{}, nil
	}
	var rows []Permission
	err := r.db.Where("code IN ?", codes).Find(&rows).Error
	return rows, err
}

func (r *Repository) ReplacePermissions(role *Role, perms []Permission) error {
	return r.db.Model(role).Association("Permissions").Replace(perms)
}

func (r *Repository) CountUsersByRole(roleID uuid.UUID) (int64, error) {
	var n int64
	err := r.db.Table("users").
		Where("role_id = ? AND deleted_at IS NULL", roleID).
		Count(&n).Error
	return n, err
}

func (r *Repository) GetRoleByID(id uuid.UUID) (*Role, error) {
	var role Role
	err := r.db.Preload("Permissions").First(&role, "id = ?", id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, apperr.ErrNotFound.WithMessage("role not found")
	}
	return &role, err
}

func (r *Repository) GetRoleByCode(code string) (*Role, error) {
	var role Role
	err := r.db.Preload("Permissions").Where("code = ?", code).First(&role).Error
	if err == gorm.ErrRecordNotFound {
		return nil, apperr.ErrNotFound.WithMessage("role not found")
	}
	return &role, err
}

func (r *Repository) CountActiveSuperAdmins() (int64, error) {
	var n int64
	err := r.db.Table("users").
		Joins("JOIN roles ON roles.id = users.role_id").
		Where("roles.code = ? AND users.is_active = ? AND users.deleted_at IS NULL", perm.RoleSuperAdmin, true).
		Count(&n).Error
	return n, err
}

func (r *Repository) ListUserIDsWithPermission(code string) ([]uuid.UUID, error) {
	var ids []uuid.UUID
	err := r.db.Table("users").
		Select("users.id").
		Joins("JOIN roles ON roles.id = users.role_id").
		Joins("JOIN role_permissions ON role_permissions.role_id = roles.id").
		Joins("JOIN permissions ON permissions.id = role_permissions.permission_id").
		Where("permissions.code = ? AND users.is_active = ? AND users.deleted_at IS NULL", code, true).
		Find(&ids).Error
	return ids, err
}
