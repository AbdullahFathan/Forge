package server_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"workspace/internal/auth"
	"workspace/internal/rbac/perm"
	"workspace/pkg/middleware"
	"workspace/pkg/response"
)

func TestAuthAndRBAC(t *testing.T) {
	secret := "test-secret-at-least-32-characters-long"
	uid := uuid.New()
	token, err := auth.SignAccess(secret, uid, perm.RoleMember, []string{perm.TaskManage}, time.Hour)
	require.NoError(t, err)

	r := chi.NewRouter()
	r.Group(func(r chi.Router) {
		r.Use(middleware.Authenticate(secret))
		r.Get("/me", func(w http.ResponseWriter, r *http.Request) {
			response.JSON(w, 200, map[string]string{"ok": "1"})
		})
		r.With(middleware.Require(perm.UserManage)).Get("/admin", func(w http.ResponseWriter, r *http.Request) {
			response.JSON(w, 200, map[string]string{"ok": "1"})
		})
	})

	t.Run("no token 401", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/me", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		require.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("valid token 200", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/me", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("missing permission 403", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/admin", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		require.Equal(t, http.StatusForbidden, rec.Code)
	})

	t.Run("admin token allowed", func(t *testing.T) {
		adminTok, err := auth.SignAccess(secret, uid, perm.RoleAdmin, []string{perm.UserManage}, time.Hour)
		require.NoError(t, err)
		req := httptest.NewRequest(http.MethodGet, "/admin", nil)
		req.Header.Set("Authorization", "Bearer "+adminTok)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Code)
	})
}
