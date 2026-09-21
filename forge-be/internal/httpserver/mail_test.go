package server

import (
	"testing"

	"github.com/stretchr/testify/require"

	"workspace/config"
	"workspace/internal/notification"
)

func TestMailSenderNopWithoutHost(t *testing.T) {
	s := mailSender(&config.Config{}, nil)
	_, ok := s.(notification.NopSender)
	require.True(t, ok)
}

func TestMailSenderSMTPWhenHostSet(t *testing.T) {
	s := mailSender(&config.Config{SMTPHost: "smtp.example.com", SMTPFrom: "a@b.com"}, nil)
	_, ok := s.(*notification.SMTPSender)
	require.True(t, ok)
}
