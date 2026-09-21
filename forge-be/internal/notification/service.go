package notification

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"

	"workspace/pkg/apperr"
)

type Sender interface {
	Send(ctx context.Context, userID uuid.UUID, ev Event) error
}

type NopSender struct{}

func (NopSender) Send(context.Context, uuid.UUID, Event) error { return nil }

type EmailPref interface {
	EmailNotificationsEnabled(userID uuid.UUID) (bool, error)
}

type Service struct {
	db     *gorm.DB
	mail   Sender
	prefs  EmailPref
	mailed int
}

func NewService(db *gorm.DB, mail Sender, prefs EmailPref) *Service {
	if mail == nil {
		mail = NopSender{}
	}
	return &Service{db: db, mail: mail, prefs: prefs}
}

func (s *Service) Emit(ctx context.Context, ev Event) error {
	if ev.UserID == uuid.Nil || ev.Type == "" || ev.EntityID == uuid.Nil {
		return nil
	}
	if ev.Day.IsZero() {
		ev.Day = time.Now().UTC()
	}
	n := Notification{
		UserID:   ev.UserID,
		Type:     ev.Type,
		Title:    ev.Title,
		Body:     ev.Body,
		Metadata: ev.Metadata,
		EntityID: ev.EntityID,
		Day:      DateUTC(ev.Day),
	}
	err := s.db.Create(&n).Error
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil
		}
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil
		}
		return err
	}
	if s.prefs != nil {
		ok, perr := s.prefs.EmailNotificationsEnabled(ev.UserID)
		if perr == nil && ok {
			s.mailed++
			_ = s.mail.Send(ctx, ev.UserID, ev)
		}
	}
	return nil
}

func (s *Service) MailedCount() int { return s.mailed }

func (s *Service) List(f ListFilter) ([]Notification, int64, error) {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.PageSize < 1 || f.PageSize > 100 {
		f.PageSize = 20
	}
	q := s.db.Model(&Notification{}).Where("user_id = ?", f.UserID)
	if f.Unread != nil {
		q = q.Where("is_read = ?", !*f.Unread)
	}
	if f.Type != "" {
		q = q.Where("type = ?", f.Type)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []Notification
	err := q.Order("created_at DESC").
		Offset((f.Page - 1) * f.PageSize).
		Limit(f.PageSize).
		Find(&rows).Error
	return rows, total, err
}

func (s *Service) UnreadCount(userID uuid.UUID) (int64, error) {
	var n int64
	err := s.db.Model(&Notification{}).Where("user_id = ? AND is_read = ?", userID, false).Count(&n).Error
	return n, err
}

func (s *Service) MarkRead(userID, id uuid.UUID) error {
	res := s.db.Model(&Notification{}).Where("id = ? AND user_id = ?", id, userID).Update("is_read", true)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return apperr.ErrNotFound.WithMessage("notification not found")
	}
	return nil
}

func (s *Service) MarkAllRead(userID uuid.UUID) error {
	return s.db.Model(&Notification{}).Where("user_id = ? AND is_read = ?", userID, false).Update("is_read", true).Error
}

func (s *Service) Latest(userID uuid.UUID, limit int) ([]Notification, error) {
	if limit < 1 {
		limit = 5
	}
	var rows []Notification
	err := s.db.Where("user_id = ?", userID).Order("created_at DESC").Limit(limit).Find(&rows).Error
	return rows, err
}
