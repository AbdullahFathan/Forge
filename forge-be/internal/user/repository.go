package user

import (
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"workspace/pkg/apperr"
)

type ListFilter struct {
	Page         int
	PageSize     int
	DepartmentID *uuid.UUID
	RoleID       *uuid.UUID
}

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) preload() *gorm.DB {
	return r.db.Preload("Role.Permissions").Preload("Department").Preload("Skills")
}

func (r *Repository) Create(u *User) error {
	return r.db.Create(u).Error
}

func (r *Repository) Save(u *User) error {
	return r.db.Save(u).Error
}

func (r *Repository) GetByID(id uuid.UUID) (*User, error) {
	var u User
	err := r.preload().First(&u, "id = ?", id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, apperr.ErrNotFound.WithMessage("user not found")
	}
	return &u, err
}

func (r *Repository) GetByEmail(email string) (*User, error) {
	var u User
	err := r.preload().Where("LOWER(email) = ?", strings.ToLower(email)).First(&u).Error
	if err == gorm.ErrRecordNotFound {
		return nil, apperr.ErrNotFound.WithMessage("user not found")
	}
	return &u, err
}

func (r *Repository) EmailTaken(email string, excludeID *uuid.UUID) (bool, error) {
	q := r.db.Model(&User{}).Where("LOWER(email) = ?", strings.ToLower(email))
	if excludeID != nil {
		q = q.Where("id <> ?", *excludeID)
	}
	var n int64
	if err := q.Count(&n).Error; err != nil {
		return false, err
	}
	return n > 0, nil
}

func (r *Repository) List(f ListFilter) ([]User, int64, error) {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.PageSize < 1 || f.PageSize > 100 {
		f.PageSize = 20
	}
	q := r.preload().Model(&User{})
	if f.DepartmentID != nil {
		q = q.Where("department_id = ?", *f.DepartmentID)
	}
	if f.RoleID != nil {
		q = q.Where("role_id = ?", *f.RoleID)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []User
	err := q.Order("created_at DESC").
		Offset((f.Page - 1) * f.PageSize).
		Limit(f.PageSize).
		Find(&rows).Error
	return rows, total, err
}

func (r *Repository) ReplaceSkills(userID uuid.UUID, skills []string) error {
	seen := map[string]struct{}{}
	rows := make([]Skill, 0, len(skills))
	for _, raw := range skills {
		s := strings.ToLower(strings.TrimSpace(raw))
		if s == "" {
			continue
		}
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		rows = append(rows, Skill{UserID: userID, Skill: s})
	}
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ?", userID).Delete(&Skill{}).Error; err != nil {
			return err
		}
		if len(rows) == 0 {
			return nil
		}
		return tx.Create(&rows).Error
	})
}

func (r *Repository) SoftDelete(id uuid.UUID) error {
	res := r.db.Delete(&User{}, "id = ?", id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return apperr.ErrNotFound.WithMessage("user not found")
	}
	return nil
}
