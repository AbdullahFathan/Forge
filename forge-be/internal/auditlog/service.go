package auditlog

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"io"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

func (s *Service) Record(_ context.Context, rec Record) error {
	if s == nil || s.db == nil {
		return nil
	}
	row := Log{
		ID:         uuid.New(),
		UserID:     rec.ActorID,
		IPAddress:  rec.IP,
		EntityType: rec.EntityType,
		EntityID:   rec.EntityID,
		ProjectID:  rec.ProjectID,
		Action:     rec.Action,
		CreatedAt:  time.Now().UTC(),
	}
	if rec.Before != nil {
		row.Before = marshal(rec.Before)
	}
	if rec.After != nil {
		row.After = marshal(rec.After)
	}
	return s.db.Create(&row).Error
}

func marshal(v any) json.RawMessage {
	if v == nil {
		return nil
	}
	if raw, ok := v.(json.RawMessage); ok {
		return raw
	}
	b, err := json.Marshal(v)
	if err != nil {
		return nil
	}
	return b
}

func (s *Service) List(f ListFilter) ([]Log, int64, error) {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.PageSize < 1 || f.PageSize > 200 {
		f.PageSize = 20
	}
	q := s.db.Model(&Log{})
	if f.From != nil {
		q = q.Where("audit_logs.created_at >= ?", f.From.UTC())
	}
	if f.To != nil {
		q = q.Where("audit_logs.created_at <= ?", f.To.UTC())
	}
	if f.UserID != nil {
		q = q.Where("audit_logs.user_id = ?", *f.UserID)
	}
	if name := strings.TrimSpace(f.UserName); name != "" {
		q = q.Joins("JOIN users ON users.id = audit_logs.user_id").
			Where("users.name ILIKE ? ESCAPE '\\'", likeContains(name))
	}
	if f.EntityType != "" {
		q = q.Where("audit_logs.entity_type = ?", f.EntityType)
	}
	if f.Action != "" {
		q = q.Where("audit_logs.action = ?", f.Action)
	}
	if f.ProjectID != nil {
		q = q.Where("audit_logs.project_id = ?", *f.ProjectID)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []Log
	err := q.Order("audit_logs.created_at DESC").
		Offset((f.Page - 1) * f.PageSize).
		Limit(f.PageSize).
		Find(&rows).Error
	return rows, total, err
}

func (s *Service) RecentForProject(projectID uuid.UUID, limit int) ([]Log, error) {
	if limit < 1 || limit > 50 {
		limit = 20
	}
	var rows []Log
	err := s.db.Where("project_id = ?", projectID).
		Order("created_at DESC").
		Limit(limit).
		Find(&rows).Error
	return rows, err
}

func WriteCSV(w io.Writer, rows []Public) error {
	cw := csv.NewWriter(w)
	if err := cw.Write([]string{"created_at", "user_name", "ip_address", "entity_type", "entity_name", "action", "before", "after"}); err != nil {
		return err
	}
	for _, r := range rows {
		if err := cw.Write([]string{
			r.CreatedAt.UTC().Format(time.RFC3339),
			r.UserName,
			r.IPAddress,
			r.EntityType,
			r.EntityName,
			r.Action,
			string(r.Before),
			string(r.After),
		}); err != nil {
			return err
		}
	}
	cw.Flush()
	return cw.Error()
}

func likeContains(value string) string {
	value = strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(value)
	return "%" + value + "%"
}
