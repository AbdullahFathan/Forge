package task

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"workspace/internal/project"
	"workspace/internal/user"
	"workspace/pkg/apperr"
)

type ListFilter struct {
	ProjectID uuid.UUID
	Page      int
	PageSize  int
	Status    string
	Assignee  *uuid.UUID
	Priority  string
	DueFrom   *time.Time
	DueTo     *time.Time
	Label     string
	Include   string
}

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) preload(includeSubtasks bool) *gorm.DB {
	q := r.db.Preload("Assignees").
		Preload("DependsOn.DependsOn").
		Preload("Dependents")
	if includeSubtasks {
		q = q.Preload("Subtasks.Assignees")
	}
	return q
}

func (r *Repository) Create(t *Task, assigneeIDs []uuid.UUID) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(t).Error; err != nil {
			return err
		}
		return r.replaceAssignees(tx, t, assigneeIDs)
	})
}

func (r *Repository) Save(t *Task, assigneeIDs *[]uuid.UUID) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(t).Error; err != nil {
			return err
		}
		if assigneeIDs != nil {
			return r.replaceAssignees(tx, t, *assigneeIDs)
		}
		return nil
	})
}

func (r *Repository) replaceAssignees(tx *gorm.DB, t *Task, ids []uuid.UUID) error {
	if err := tx.Model(t).Association("Assignees").Clear(); err != nil {
		return err
	}
	if len(ids) == 0 {
		return nil
	}
	users := make([]user.User, 0, len(ids))
	for _, id := range ids {
		users = append(users, user.User{ID: id})
	}
	return tx.Model(t).Association("Assignees").Append(users)
}

func (r *Repository) GetByID(id uuid.UUID, includeSubtasks bool) (*Task, error) {
	var t Task
	err := r.preload(includeSubtasks).First(&t, "id = ?", id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, apperr.ErrNotFound.WithMessage("task not found")
	}
	return &t, err
}

func (r *Repository) List(f ListFilter) ([]Task, int64, error) {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.PageSize < 1 || f.PageSize > 100 {
		f.PageSize = 20
	}
	include := f.Include == "subtasks"
	q := r.preload(include).Model(&Task{}).Where("project_id = ?", f.ProjectID)
	if include {
		q = q.Where("parent_task_id IS NULL")
	}
	if f.Status != "" {
		q = q.Where("status = ?", f.Status)
	}
	if f.Priority != "" {
		q = q.Where("priority = ?", f.Priority)
	}
	if f.Assignee != nil {
		q = q.Where("id IN (SELECT task_id FROM task_assignees WHERE user_id = ?)", *f.Assignee)
	}
	if f.DueFrom != nil {
		q = q.Where("due_date >= ?", *f.DueFrom)
	}
	if f.DueTo != nil {
		q = q.Where("due_date <= ?", *f.DueTo)
	}
	if f.Label != "" {
		b, _ := json.Marshal([]string{f.Label})
		q = q.Where("labels @> ?", string(b))
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []Task
	err := q.Order("status ASC, position ASC, created_at ASC").
		Offset((f.Page - 1) * f.PageSize).
		Limit(f.PageSize).
		Find(&rows).Error
	return rows, total, err
}

func (r *Repository) SoftDelete(id uuid.UUID) error {
	res := r.db.Delete(&Task{}, "id = ?", id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return apperr.ErrNotFound.WithMessage("task not found")
	}
	return nil
}

func (r *Repository) HasChildren(id uuid.UUID) (bool, error) {
	var n int64
	err := r.db.Model(&Task{}).Where("parent_task_id = ?", id).Count(&n).Error
	return n > 0, err
}

func (r *Repository) MaxPosition(projectID uuid.UUID, status string) (int, error) {
	var pos int
	err := r.db.Model(&Task{}).
		Where("project_id = ? AND status = ?", projectID, status).
		Select("COALESCE(MAX(position), 0)").
		Scan(&pos).Error
	return pos, err
}

func (r *Repository) AddDependency(d *Dependency) error {
	return r.db.Create(d).Error
}

func (r *Repository) DeleteDependency(taskID, depID uuid.UUID) error {
	res := r.db.Where("id = ? AND task_id = ?", depID, taskID).Delete(&Dependency{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return apperr.ErrNotFound.WithMessage("dependency not found")
	}
	return nil
}

func (r *Repository) DependencyEdges(projectID uuid.UUID) (map[uuid.UUID][]uuid.UUID, error) {
	var rows []Dependency
	err := r.db.Table("task_dependencies").
		Joins("JOIN tasks t ON t.id = task_dependencies.task_id AND t.deleted_at IS NULL").
		Where("t.project_id = ?", projectID).
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	edges := map[uuid.UUID][]uuid.UUID{}
	for _, row := range rows {
		edges[row.TaskID] = append(edges[row.TaskID], row.DependsOnTaskID)
	}
	return edges, nil
}

func (r *Repository) UnfinishedPredecessors(taskID uuid.UUID) (bool, error) {
	var n int64
	err := r.db.Table("task_dependencies").
		Joins("JOIN tasks pred ON pred.id = task_dependencies.depends_on_task_id AND pred.deleted_at IS NULL").
		Where("task_dependencies.task_id = ? AND pred.status <> ?", taskID, StatusDone).
		Count(&n).Error
	return n > 0, err
}

func (r *Repository) AddComment(c *Comment) error {
	return r.db.Create(c).Error
}

func (r *Repository) ListComments(taskID uuid.UUID) ([]Comment, error) {
	var rows []Comment
	err := r.db.Preload("User").Where("task_id = ?", taskID).Order("created_at ASC").Find(&rows).Error
	return rows, err
}

func (r *Repository) Completion(projectID uuid.UUID) (float64, project.TaskCounts, error) {
	type row struct {
		Status string
		N      int64
	}
	var rows []row
	err := r.db.Model(&Task{}).
		Select("status, COUNT(*) as n").
		Where("project_id = ?", projectID).
		Group("status").
		Scan(&rows).Error
	if err != nil {
		return 0, project.TaskCounts{}, err
	}
	var c project.TaskCounts
	for _, x := range rows {
		c.Total += x.N
		switch x.Status {
		case StatusBacklog:
			c.Backlog = x.N
		case StatusTodo:
			c.Todo = x.N
		case StatusInProgress:
			c.InProgress = x.N
		case StatusInReview:
			c.InReview = x.N
		case StatusDone:
			c.Done = x.N
		case StatusBlocked:
			c.Blocked = x.N
		}
	}
	var pct float64
	if c.Total > 0 {
		pct = float64(c.Done) / float64(c.Total) * 100
	}
	return pct, c, nil
}

func (r *Repository) ListByAssignee(userID uuid.UUID) ([]Task, error) {
	var rows []Task
	err := r.preload(false).
		Where("id IN (SELECT task_id FROM task_assignees WHERE user_id = ?)", userID).
		Order("due_date ASC").
		Order("created_at ASC").
		Find(&rows).Error
	return rows, err
}

func (r *Repository) ListOpenDueBetween(from, to time.Time) ([]Task, error) {
	var rows []Task
	err := r.preload(false).
		Where("due_date > ? AND due_date <= ? AND status <> ?", from, to, StatusDone).
		Find(&rows).Error
	return rows, err
}

func (r *Repository) ListOpenOverdue(before time.Time) ([]Task, error) {
	var rows []Task
	err := r.preload(false).
		Where("due_date < ? AND status <> ?", before, StatusDone).
		Find(&rows).Error
	return rows, err
}

type Attention struct {
	Blocked    []Task
	Overdue    []Task
	Unassigned []Task
}

func (r *Repository) Attention(projectIDs []uuid.UUID, today time.Time) (Attention, error) {
	var out Attention
	if len(projectIDs) == 0 {
		return out, nil
	}
	q := r.preload(false).Where("project_id IN ?", projectIDs)
	if err := q.Where("status = ?", StatusBlocked).Find(&out.Blocked).Error; err != nil {
		return out, err
	}
	if err := r.preload(false).Where("project_id IN ? AND due_date < ? AND status <> ?", projectIDs, today, StatusDone).Find(&out.Overdue).Error; err != nil {
		return out, err
	}
	if err := r.db.Where("project_id IN ? AND id NOT IN (SELECT task_id FROM task_assignees)", projectIDs).Find(&out.Unassigned).Error; err != nil {
		return out, err
	}
	return out, nil
}

func (r *Repository) CompletionByProject(projectIDs []uuid.UUID) (map[uuid.UUID]struct {
	Done  int64
	Total int64
}, error) {
	type row struct {
		ProjectID uuid.UUID
		Done      int64
		Total     int64
	}
	q := r.db.Model(&Task{}).Select("project_id, COUNT(*) FILTER (WHERE status = ?) as done, COUNT(*) as total", StatusDone)
	if len(projectIDs) > 0 {
		q = q.Where("project_id IN ?", projectIDs)
	}
	var rows []row
	err := q.Group("project_id").Scan(&rows).Error
	out := map[uuid.UUID]struct {
		Done  int64
		Total int64
	}{}
	for _, x := range rows {
		out[x.ProjectID] = struct {
			Done  int64
			Total int64
		}{Done: x.Done, Total: x.Total}
	}
	return out, err
}

func (r *Repository) CompletionByUser(projectIDs []uuid.UUID) (map[uuid.UUID]struct {
	Done  int64
	Total int64
}, error) {
	type row struct {
		UserID uuid.UUID
		Done   int64
		Total  int64
	}
	q := r.db.Table("tasks").
		Select("task_assignees.user_id as user_id, COUNT(*) FILTER (WHERE tasks.status = ?) as done, COUNT(*) as total", StatusDone).
		Joins("JOIN task_assignees ON task_assignees.task_id = tasks.id").
		Where("tasks.deleted_at IS NULL")
	if len(projectIDs) > 0 {
		q = q.Where("tasks.project_id IN ?", projectIDs)
	}
	var rows []row
	err := q.Group("task_assignees.user_id").Scan(&rows).Error
	out := map[uuid.UUID]struct {
		Done  int64
		Total int64
	}{}
	for _, x := range rows {
		out[x.UserID] = struct {
			Done  int64
			Total int64
		}{Done: x.Done, Total: x.Total}
	}
	return out, err
}

func (r *Repository) IsAssignee(taskID, userID uuid.UUID) (bool, error) {
	var n int64
	err := r.db.Table("task_assignees").Where("task_id = ? AND user_id = ?", taskID, userID).Count(&n).Error
	return n > 0, err
}
