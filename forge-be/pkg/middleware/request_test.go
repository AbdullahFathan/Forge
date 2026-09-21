package middleware

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestRequestLoggerOmitsSecrets(t *testing.T) {
	core, logs := observer.New(zapcore.InfoLevel)
	log := zap.New(core)
	h := RequestLogger(log)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest(http.MethodGet, "/users/me", nil)
	req.Header.Set("Authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.payload.sig")
	req.Header.Set("Cookie", "refresh_token=super-secret-refresh")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, 1, logs.Len())
	entry := logs.All()[0]
	blob := fmt.Sprintf("%s %v", entry.Message, entry.ContextMap())
	require.NotContains(t, blob, "Bearer ")
	require.NotContains(t, blob, "eyJhbGci")
	require.NotContains(t, blob, "refresh_token")
	require.NotContains(t, blob, "super-secret-refresh")
	require.Equal(t, "/users/me", entry.ContextMap()["path"])
}
