package user

import (
	"errors"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"workspace/internal/department"
	"workspace/internal/rbac"
	"workspace/internal/rbac/perm"
	"workspace/pkg/apperr"
)

type RoleFinder interface {
	GetRoleByID(id uuid.UUID) (*rbac.Role, error)
	CountActiveSuperAdmins() (int64, error)
}

type DeptFinder interface {
	GetByID(id uuid.UUID) (*department.Department, error)
}

type Service struct {
	users *Repository
	roles RoleFinder
	depts DeptFinder
}

func NewService(users *Repository, roles RoleFinder, depts DeptFinder) *Service {
	return &Service{users: users, roles: roles, depts: depts}
}

type CreateInput struct {
	Name                string
	Email               string
	Password            string
	RoleID              uuid.UUID
	DepartmentID        *uuid.UUID
	CapacityHoursPerDay int
	IsActive            bool
}

type PatchInput struct {
	Name                *string
	Password            *string
	RoleID              *uuid.UUID
	DepartmentID        **uuid.UUID
	CapacityHoursPerDay *int
	IsActive            *bool
}

type MePatchInput struct {
	Name            *string
	Password        *string
	CurrentPassword *string
}

func (s *Service) Get(id uuid.UUID) (*User, error) {
	return s.users.GetByID(id)
}

func (s *Service) List(f ListFilter) ([]User, int64, error) {
	return s.users.List(f)
}

func (s *Service) Create(in CreateInput) (*User, error) {
	taken, err := s.users.EmailTaken(in.Email, nil)
	if err != nil {
		return nil, err
	}
	if taken {
		return nil, apperr.ErrConflict.WithMessage("email already in use")
	}
	if _, err := s.roles.GetRoleByID(in.RoleID); err != nil {
		return nil, apperr.ErrValidation.WithMessage("invalid roleId")
	}
	if in.DepartmentID != nil {
		if _, err := s.depts.GetByID(*in.DepartmentID); err != nil {
			return nil, apperr.ErrValidation.WithMessage("invalid departmentId")
		}
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), 12)
	if err != nil {
		return nil, apperr.ErrInternal
	}
	cap := in.CapacityHoursPerDay
	if cap == 0 {
		cap = 8
	}
	u := &User{
		Name:                in.Name,
		Email:               in.Email,
		PasswordHash:        string(hash),
		RoleID:              in.RoleID,
		DepartmentID:        in.DepartmentID,
		CapacityHoursPerDay: cap,
		IsActive:            in.IsActive,
	}
	if err := s.users.Create(u); err != nil {
		return nil, err
	}
	return s.users.GetByID(u.ID)
}

func (s *Service) Patch(id uuid.UUID, in PatchInput) (*User, error) {
	u, err := s.users.GetByID(id)
	if err != nil {
		return nil, err
	}
	if err := s.guardLastSuperAdmin(u, in); err != nil {
		return nil, err
	}
	if in.Name != nil {
		u.Name = *in.Name
	}
	if in.Password != nil {
		hash, err := bcrypt.GenerateFromPassword([]byte(*in.Password), 12)
		if err != nil {
			return nil, apperr.ErrInternal
		}
		u.PasswordHash = string(hash)
	}
	if in.RoleID != nil {
		if _, err := s.roles.GetRoleByID(*in.RoleID); err != nil {
			return nil, apperr.ErrValidation.WithMessage("invalid roleId")
		}
		u.RoleID = *in.RoleID
	}
	if in.DepartmentID != nil {
		if *in.DepartmentID != nil {
			if _, err := s.depts.GetByID(**in.DepartmentID); err != nil {
				return nil, apperr.ErrValidation.WithMessage("invalid departmentId")
			}
		}
		u.DepartmentID = *in.DepartmentID
	}
	if in.CapacityHoursPerDay != nil {
		u.CapacityHoursPerDay = *in.CapacityHoursPerDay
	}
	if in.IsActive != nil {
		u.IsActive = *in.IsActive
	}
	if err := s.users.Save(u); err != nil {
		return nil, err
	}
	return s.users.GetByID(u.ID)
}

func (s *Service) PatchMe(id uuid.UUID, in MePatchInput) (*User, error) {
	u, err := s.users.GetByID(id)
	if err != nil {
		return nil, err
	}
	if in.Name != nil {
		u.Name = *in.Name
	}
	if in.Password != nil {
		if in.CurrentPassword == nil {
			return nil, apperr.ErrValidation.WithMessage("currentPassword is required")
		}
		if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(*in.CurrentPassword)) != nil {
			return nil, apperr.ErrInvalidCredentials.WithMessage("current password is incorrect")
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(*in.Password), 12)
		if err != nil {
			return nil, apperr.ErrInternal
		}
		u.PasswordHash = string(hash)
	}
	if err := s.users.Save(u); err != nil {
		return nil, err
	}
	return s.users.GetByID(u.ID)
}

func (s *Service) Delete(id uuid.UUID) error {
	u, err := s.users.GetByID(id)
	if err != nil {
		return err
	}
	n, err := s.roles.CountActiveSuperAdmins()
	if err != nil {
		return err
	}
	if err := GuardLastSuperAdmin(u.Role.Code, true, n); err != nil {
		return err
	}
	return s.users.SoftDelete(id)
}

func GuardLastSuperAdmin(roleCode string, removing bool, activeCount int64) error {
	if roleCode != perm.RoleSuperAdmin || !removing {
		return nil
	}
	if activeCount <= 1 {
		return apperr.ErrLastSuperAdmin
	}
	return nil
}

func (s *Service) guardLastSuperAdmin(u *User, in PatchInput) error {
	demote := false
	if in.RoleID != nil && *in.RoleID != u.RoleID {
		demote = true
	}
	if in.IsActive != nil && !*in.IsActive {
		demote = true
	}
	n, err := s.roles.CountActiveSuperAdmins()
	if err != nil {
		return err
	}
	return GuardLastSuperAdmin(u.Role.Code, demote, n)
}

func IsNotFound(err error) bool {
	if err == nil {
		return false
	}
	ae, ok := apperr.As(err)
	if ok && ae.Code == apperr.ErrNotFound.Code {
		return true
	}
	return errors.Is(err, apperr.ErrNotFound)
}
