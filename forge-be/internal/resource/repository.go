package resource

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"workspace/internal/user"
	"workspace/pkg/apperr"
)

type ListFilter struct {
	Page      int
	PageSize  int
	UserID    *uuid.UUID
	ProjectID *uuid.UUID
	From      *time.Time
	To        *time.Time
	ScopeAll  bool
	ActorID   uuid.UUID
}

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) preload() *gorm.DB {
	return r.db.Preload("User.Department").Preload("Project")
}

func (r *Repository) Create(a *Allocation) error {
	return r.db.Create(a).Error
}

func (r *Repository) Save(a *Allocation) error {
	return r.db.Save(a).Error
}

func (r *Repository) GetByID(id uuid.UUID) (*Allocation, error) {
	var a Allocation
	err := r.preload().First(&a, "id = ?", id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, apperr.ErrNotFound.WithMessage("allocation not found")
	}
	return &a, err
}

func (r *Repository) SoftDelete(id uuid.UUID) error {
	res := r.db.Delete(&Allocation{}, "id = ?", id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return apperr.ErrNotFound.WithMessage("allocation not found")
	}
	return nil
}

func (r *Repository) applyScope(q *gorm.DB, f ListFilter) *gorm.DB {
	if f.ScopeAll {
		return q
	}
	return q.Where(
		`project_id IN (
			SELECT id FROM projects WHERE owner_id = ? AND deleted_at IS NULL
			UNION
			SELECT project_id FROM project_members WHERE user_id = ? AND role = ?
		)`,
		f.ActorID, f.ActorID, RoleLead,
	)
}

func (r *Repository) List(f ListFilter) ([]Allocation, int64, error) {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.PageSize < 1 || f.PageSize > 100 {
		f.PageSize = 20
	}
	q := r.preload().Model(&Allocation{})
	q = r.applyScope(q, f)
	if f.UserID != nil {
		q = q.Where("user_id = ?", *f.UserID)
	}
	if f.ProjectID != nil {
		q = q.Where("project_id = ?", *f.ProjectID)
	}
	if f.From != nil && f.To != nil {
		q = q.Where("start_date <= ? AND end_date >= ?", DateUTC(*f.To), DateUTC(*f.From))
	} else if f.From != nil {
		q = q.Where("end_date >= ?", DateUTC(*f.From))
	} else if f.To != nil {
		q = q.Where("start_date <= ?", DateUTC(*f.To))
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []Allocation
	err := q.Order("start_date ASC, created_at ASC").
		Offset((f.Page - 1) * f.PageSize).
		Limit(f.PageSize).
		Find(&rows).Error
	return rows, total, err
}

func (r *Repository) OverlappingSameProject(userID, projectID uuid.UUID, from, to time.Time, excludeID *uuid.UUID) ([]Allocation, error) {
	q := r.db.Model(&Allocation{}).
		Where("user_id = ? AND project_id = ?", userID, projectID).
		Where("start_date <= ? AND end_date >= ?", DateUTC(to), DateUTC(from))
	if excludeID != nil {
		q = q.Where("id <> ?", *excludeID)
	}
	var rows []Allocation
	err := q.Find(&rows).Error
	return rows, err
}

func (r *Repository) ForUsersInRange(userIDs []uuid.UUID, from, to time.Time) ([]Allocation, error) {
	if len(userIDs) == 0 {
		return nil, nil
	}
	var rows []Allocation
	err := r.preload().
		Where("user_id IN ?", userIDs).
		Where("start_date <= ? AND end_date >= ?", DateUTC(to), DateUTC(from)).
		Find(&rows).Error
	return rows, err
}

func (r *Repository) AllInRange(from, to time.Time) ([]Allocation, error) {
	var rows []Allocation
	err := r.preload().
		Where("start_date <= ? AND end_date >= ?", DateUTC(to), DateUTC(from)).
		Find(&rows).Error
	return rows, err
}

func (r *Repository) ListHolidays() ([]Holiday, error) {
	var rows []Holiday
	err := r.db.Order("date ASC").Find(&rows).Error
	return rows, err
}

func (r *Repository) CreateHoliday(h *Holiday) error {
	return r.db.Create(h).Error
}

func (r *Repository) GetHolidayByDate(d time.Time) (*Holiday, error) {
	var h Holiday
	err := r.db.Where("date = ?", DateUTC(d)).First(&h).Error
	if err == gorm.ErrRecordNotFound {
		return nil, apperr.ErrNotFound.WithMessage("holiday not found")
	}
	return &h, err
}

func (r *Repository) DeleteHoliday(id uuid.UUID) error {
	res := r.db.Delete(&Holiday{}, "id = ?", id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return apperr.ErrNotFound.WithMessage("holiday not found")
	}
	return nil
}

func (r *Repository) ListActiveUsers(departmentID *uuid.UUID, skill string) ([]user.User, error) {
	q := r.db.Preload("Department").Preload("Skills").Model(&user.User{}).Where("is_active = ?", true)
	if departmentID != nil {
		q = q.Where("department_id = ?", *departmentID)
	}
	skill = strings.ToLower(strings.TrimSpace(skill))
	if skill != "" {
		q = q.Where("id IN (SELECT user_id FROM user_skills WHERE skill = ?)", skill)
	}
	var rows []user.User
	err := q.Order("name ASC").Find(&rows).Error
	return rows, err
}

func (r *Repository) GetUser(id uuid.UUID) (*user.User, error) {
	var u user.User
	err := r.db.Preload("Department").Preload("Skills").First(&u, "id = ?", id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, apperr.ErrNotFound.WithMessage("user not found")
	}
	return &u, err
}
