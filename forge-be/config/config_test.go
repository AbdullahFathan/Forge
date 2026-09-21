package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadFailsWithoutRequiredEnv(t *testing.T) {
	loadDotenv = false
	t.Cleanup(func() { loadDotenv = true })
	keys := []string{
		"DATABASE_DSN", "REDIS_ADDR", "JWT_SECRET",
		"BOOTSTRAP_ADMIN_EMAIL", "BOOTSTRAP_ADMIN_PASSWORD",
	}
	for _, k := range keys {
		t.Setenv(k, "")
	}
	_, err := Load()
	require.Error(t, err)
}

func TestLoadOK(t *testing.T) {
	t.Setenv("DATABASE_DSN", "postgres://u:p@localhost:5432/db?sslmode=disable")
	t.Setenv("REDIS_ADDR", "localhost:6379")
	t.Setenv("JWT_SECRET", "test-secret-at-least-32-characters-long")
	t.Setenv("BOOTSTRAP_ADMIN_EMAIL", "a@b.com")
	t.Setenv("BOOTSTRAP_ADMIN_PASSWORD", "password12")
	t.Setenv("CORS_ORIGINS", "http://localhost:3000")
	cfg, err := Load()
	require.NoError(t, err)
	require.Equal(t, "8080", cfg.HTTPPort)
	require.Equal(t, 25, cfg.DBMaxOpenConns)
	require.Equal(t, "587", cfg.SMTPPort)
}
