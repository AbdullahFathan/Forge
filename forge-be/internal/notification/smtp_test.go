package notification

import (
	"context"
	"net/smtp"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"workspace/internal/user"
)

type memUsers struct {
	u *user.User
}

func (m memUsers) GetByID(id uuid.UUID) (*user.User, error) {
	return m.u, nil
}

func TestFormatMessage(t *testing.T) {
	msg := string(FormatMessage("from@x.com", "to@x.com", "Hello\nWorld", "body line"))
	require.Contains(t, msg, "From: from@x.com")
	require.Contains(t, msg, "To: to@x.com")
	require.Contains(t, msg, "Subject: Hello World")
	require.Contains(t, msg, "body line")
	require.NotContains(t, msg, "Subject: Hello\n")
}

func TestSMTPSenderSend(t *testing.T) {
	uid := uuid.New()
	s := NewSMTP(SMTPConfig{Host: "smtp.example", Port: "587", From: "noreply@x.com", Username: "u", Password: "p"}, memUsers{&user.User{ID: uid, Email: "a@b.com"}})
	var gotAddr string
	var gotTo []string
	s.send = func(addr string, a smtp.Auth, from string, to []string, msg []byte) error {
		gotAddr = addr
		gotTo = to
		require.Equal(t, "noreply@x.com", from)
		require.Contains(t, string(msg), "assigned")
		return nil
	}
	err := s.Send(context.Background(), uid, Event{Title: "Task assigned", Body: "You have a task"})
	require.NoError(t, err)
	require.Equal(t, "smtp.example:587", gotAddr)
	require.Equal(t, []string{"a@b.com"}, gotTo)
}
