package notification

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"net/smtp"
	"strings"

	"github.com/google/uuid"

	"workspace/internal/user"
)

type UserEmail interface {
	GetByID(id uuid.UUID) (*user.User, error)
}

type SMTPConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	From     string
	StartTLS bool
}

type SMTPSender struct {
	cfg   SMTPConfig
	users UserEmail
	send  func(addr string, a smtp.Auth, from string, to []string, msg []byte) error
}

func NewSMTP(cfg SMTPConfig, users UserEmail) *SMTPSender {
	return &SMTPSender{cfg: cfg, users: users, send: smtp.SendMail}
}

func FormatMessage(from, to, subject, body string) []byte {
	var b bytes.Buffer
	fmt.Fprintf(&b, "From: %s\r\n", from)
	fmt.Fprintf(&b, "To: %s\r\n", to)
	fmt.Fprintf(&b, "Subject: %s\r\n", sanitizeHeader(subject))
	fmt.Fprintf(&b, "MIME-Version: 1.0\r\n")
	fmt.Fprintf(&b, "Content-Type: text/plain; charset=UTF-8\r\n\r\n")
	b.WriteString(body)
	if !strings.HasSuffix(body, "\n") {
		b.WriteString("\r\n")
	}
	return b.Bytes()
}

func sanitizeHeader(s string) string {
	s = strings.ReplaceAll(s, "\r", " ")
	s = strings.ReplaceAll(s, "\n", " ")
	return strings.TrimSpace(s)
}

func (s *SMTPSender) Send(ctx context.Context, userID uuid.UUID, ev Event) error {
	_ = ctx
	if s.users == nil {
		return nil
	}
	u, err := s.users.GetByID(userID)
	if err != nil {
		return err
	}
	from := s.cfg.From
	if from == "" {
		from = s.cfg.Username
	}
	if from == "" || u.Email == "" {
		return nil
	}
	_ = s.cfg.StartTLS // smtp.SendMail upgrades when the server advertises STARTTLS (typical on 587)
	port := s.cfg.Port
	if port == "" {
		port = "587"
	}
	addr := net.JoinHostPort(s.cfg.Host, port)
	var auth smtp.Auth
	if s.cfg.Username != "" {
		auth = smtp.PlainAuth("", s.cfg.Username, s.cfg.Password, s.cfg.Host)
	}
	msg := FormatMessage(from, u.Email, ev.Title, ev.Body)
	return s.send(addr, auth, from, []string{u.Email}, msg)
}
