package project

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"workspace/pkg/apperr"
)

type ListFilter struct {
	Page         int
	PageSize     int
	Status       string
	DepartmentID *uuid.UUID
	ScopeAll     bool
	UserID       uuid.UUID
}

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) preload() *gorm.DB {
	return r.db.Preload("Owner").Preload("Department")
}

func (r *Repository) Create(p *Project) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(p).Error; err != nil {
			return err
		}
		m := Member{ProjectID: p.ID, UserID: p.OwnerID, Role: RoleLead}
		return tx.Create(&m).Error
	})
}

func (r *Repository) Save(p *Project) error {
	return r.db.Save(p).Error
}

func (r *Repository) GetByID(id uuid.UUID) (*Project, error) {
	var p Project
	err := r.preload().Unscoped().First(&p, "id = ?", id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, apperr.ErrNotFound.WithMessage("project not found")
	}
	return &p, err
}

// List is served by idx_projects_status_department_id, idx_projects_owner_id,
// and idx_project_members_user_id (membership subquery). Archived lists use Unscoped.
func (r *Repository) List(f ListFilter) ([]Project, int64, error) {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.PageSize < 1 || f.PageSize > 100 {
		f.PageSize = 20
	}
	q := r.preload().Model(&Project{})
	if f.Status == StatusArchived {
		q = q.Unscoped().Where("status = ?", StatusArchived)
	} else if f.Status != "" {
		q = q.Where("status = ?", f.Status)
	}
	if f.DepartmentID != nil {
		q = q.Where("department_id = ?", *f.DepartmentID)
	}
	if !f.ScopeAll {
		q = q.Where(
			"owner_id = ? OR id IN (SELECT project_id FROM project_members WHERE user_id = ?)",
			f.UserID, f.UserID,
		)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []Project
	err := q.Order("created_at DESC").
		Offset((f.Page - 1) * f.PageSize).
		Limit(f.PageSize).
		Find(&rows).Error
	return rows, total, err
}

func (r *Repository) Archive(id uuid.UUID) error {
	p, err := r.GetByID(id)
	if err != nil {
		return err
	}
	p.Status = StatusArchived
	if err := r.db.Save(p).Error; err != nil {
		return err
	}
	res := r.db.Delete(&Project{}, "id = ?", id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return apperr.ErrNotFound.WithMessage("project not found")
	}
	return nil
}

func (r *Repository) GetMember(projectID, userID uuid.UUID) (*Member, error) {
	var m Member
	err := r.db.Preload("User.Department").
		Where("project_id = ? AND user_id = ?", projectID, userID).
		First(&m).Error
	if err == gorm.ErrRecordNotFound {
		return nil, apperr.ErrNotFound.WithMessage("member not found")
	}
	return &m, err
}

func (r *Repository) ListMembers(projectID uuid.UUID, page, pageSize int) ([]Member, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	q := r.db.Model(&Member{}).Where("project_id = ?", projectID)
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []Member
	err := r.db.Preload("User.Department").
		Where("project_id = ?", projectID).
		Order("created_at ASC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&rows).Error
	return rows, total, err
}

func (r *Repository) CountMembers(projectID uuid.UUID) (int64, error) {
	var n int64
	err := r.db.Model(&Member{}).Where("project_id = ?", projectID).Count(&n).Error
	return n, err
}

func (r *Repository) AddMember(m *Member) error {
	return r.db.Create(m).Error
}

func (r *Repository) SaveMember(m *Member) error {
	return r.db.Save(m).Error
}

func (r *Repository) UpsertMember(m *Member) error {
	existing, err := r.GetMember(m.ProjectID, m.UserID)
	if err != nil {
		if ae, ok := apperr.As(err); ok && ae.Code == apperr.ErrNotFound.Code {
			return r.AddMember(m)
		}
		return err
	}
	existing.Role = m.Role
	return r.SaveMember(existing)
}

func (r *Repository) DeleteMember(projectID, userID uuid.UUID) error {
	res := r.db.Where("project_id = ? AND user_id = ?", projectID, userID).Delete(&Member{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return apperr.ErrNotFound.WithMessage("member not found")
	}
	return nil
}

func (r *Repository) IsMember(projectID, userID uuid.UUID) (bool, string, error) {
	var m Member
	err := r.db.Select("role").Where("project_id = ? AND user_id = ?", projectID, userID).First(&m).Error
	if err == gorm.ErrRecordNotFound {
		return false, "", nil
	}
	if err != nil {
		return false, "", err
	}
	return true, m.Role, nil
}

type StatusCounts struct {
	Active    int64
	Done      int64
	Late      int64
}

func (r *Repository) StatusCounts(today time.Time) (StatusCounts, error) {
	today = today.UTC().Truncate(24 * time.Hour)
	var out StatusCounts
	err := r.db.Model(&Project{}).Where("status = ? AND deleted_at IS NULL", StatusActive).Count(&out.Active).Error
	if err != nil {
		return out, err
	}
	err = r.db.Model(&Project{}).Where("status = ?", StatusCompleted).Count(&out.Done).Error
	if err != nil {
		return out, err
	}
	err = r.db.Model(&Project{}).
		Where("target_end_date < ? AND status NOT IN ? AND deleted_at IS NULL", today, []string{StatusCompleted, StatusArchived}).
		Count(&out.Late).Error
	return out, err
}

func (r *Repository) DueBetween(from, to time.Time) ([]Project, error) {
	var rows []Project
	err := r.preload().
		Where("target_end_date >= ? AND target_end_date <= ? AND status NOT IN ? AND deleted_at IS NULL",
			from, to, []string{StatusCompleted, StatusArchived}).
		Order("target_end_date ASC").
		Find(&rows).Error
	return rows, err
}

type MonthCount struct {
	Month string
	Count int64
}

func (r *Repository) CompletedByMonth(from time.Time) ([]MonthCount, error) {
	type row struct {
		Month string
		N     int64
	}
	var rows []row
	err := r.db.Model(&Project{}).
		Select("to_char(date_trunc('month', updated_at), 'YYYY-MM') as month, count(*) as n").
		Where("status = ? AND updated_at >= ?", StatusCompleted, from).
		Group("date_trunc('month', updated_at)").
		Order("month").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]MonthCount, 0, len(rows))
	for _, x := range rows {
		out = append(out, MonthCount{Month: x.Month, Count: x.N})
	}
	return out, nil
}

func (r *Repository) ListOwned(ownerID uuid.UUID) ([]Project, error) {
	var rows []Project
	err := r.preload().Where("owner_id = ? AND deleted_at IS NULL", ownerID).Order("created_at DESC").Find(&rows).Error
	return rows, err
}

func (r *Repository) ListAllActive() ([]Project, error) {
	var rows []Project
	err := r.preload().Where("deleted_at IS NULL").Find(&rows).Error
	return rows, err
}

func (r *Repository) MemberUserIDs(projectID uuid.UUID) ([]uuid.UUID, error) {
	var ids []uuid.UUID
	err := r.db.Model(&Member{}).Where("project_id = ?", projectID).Pluck("user_id", &ids).Error
	return ids, err
}
