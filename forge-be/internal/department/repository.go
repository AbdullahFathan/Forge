package department

import (
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"workspace/pkg/apperr"
)

type ListFilter struct {
	Page     int
	PageSize int
}

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(d *Department) error {
	return r.db.Create(d).Error
}

func (r *Repository) Save(d *Department) error {
	return r.db.Save(d).Error
}

func (r *Repository) GetByID(id uuid.UUID) (*Department, error) {
	var d Department
	err := r.db.First(&d, "id = ?", id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, apperr.ErrNotFound.WithMessage("department not found")
	}
	return &d, err
}

func (r *Repository) NameTaken(name string, excludeID *uuid.UUID) (bool, error) {
	q := r.db.Model(&Department{}).Where("LOWER(name) = ?", strings.ToLower(name))
	if excludeID != nil {
		q = q.Where("id <> ?", *excludeID)
	}
	var n int64
	if err := q.Count(&n).Error; err != nil {
		return false, err
	}
	return n > 0, nil
}

func (r *Repository) List(f ListFilter) ([]Department, int64, error) {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.PageSize < 1 || f.PageSize > 100 {
		f.PageSize = 20
	}
	q := r.db.Model(&Department{})
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []Department
	err := q.Order("name ASC").
		Offset((f.Page - 1) * f.PageSize).
		Limit(f.PageSize).
		Find(&rows).Error
	return rows, total, err
}

func (r *Repository) SoftDelete(id uuid.UUID) error {
	res := r.db.Delete(&Department{}, "id = ?", id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return apperr.ErrNotFound.WithMessage("department not found")
	}
	return nil
}
