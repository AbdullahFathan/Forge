package resource_test

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

func TestResourceRouteRBAC(t *testing.T) {
	secret := "test-secret-at-least-32-characters-long"
	memberTok, err := auth.SignAccess(secret, uuid.New(), perm.RoleMember, []string{perm.TaskManage}, time.Hour)
	require.NoError(t, err)
	pmTok, err := auth.SignAccess(secret, uuid.New(), perm.RoleProjectManager, []string{perm.ResourceAllocate, perm.ProjectCreate}, time.Hour)
	require.NoError(t, err)
	rmTok, err := auth.SignAccess(secret, uuid.New(), perm.RoleResourceManager, []string{perm.ResourceAllocate, perm.CapacityView, perm.ProjectReadAll}, time.Hour)
	require.NoError(t, err)

	r := chi.NewRouter()
	r.Group(func(r chi.Router) {
		r.Use(middleware.Authenticate(secret))
		r.With(middleware.Require(perm.ResourceAllocate)).Get("/resources/people", func(w http.ResponseWriter, r *http.Request) {
			response.JSON(w, http.StatusOK, map[string]string{"ok": "1"})
		})
		r.With(middleware.Require(perm.ResourceAllocate)).Post("/resources/allocations", func(w http.ResponseWriter, r *http.Request) {
			response.JSON(w, http.StatusCreated, map[string]string{"ok": "1"})
		})
		r.With(middleware.Require(perm.CapacityView)).Get("/resources/capacity", func(w http.ResponseWriter, r *http.Request) {
			response.JSON(w, http.StatusOK, map[string]string{"ok": "1"})
		})
	})

	req := httptest.NewRequest(http.MethodGet, "/resources/people", nil)
	req.Header.Set("Authorization", "Bearer "+memberTok)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	require.Equal(t, http.StatusForbidden, rec.Code)

	req = httptest.NewRequest(http.MethodGet, "/resources/people", nil)
	req.Header.Set("Authorization", "Bearer "+pmTok)
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	req = httptest.NewRequest(http.MethodPost, "/resources/allocations", nil)
	req.Header.Set("Authorization", "Bearer "+memberTok)
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	require.Equal(t, http.StatusForbidden, rec.Code)

	req = httptest.NewRequest(http.MethodPost, "/resources/allocations", nil)
	req.Header.Set("Authorization", "Bearer "+pmTok)
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	require.Equal(t, http.StatusCreated, rec.Code)

	req = httptest.NewRequest(http.MethodGet, "/resources/capacity", nil)
	req.Header.Set("Authorization", "Bearer "+pmTok)
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	require.Equal(t, http.StatusForbidden, rec.Code)

	req = httptest.NewRequest(http.MethodGet, "/resources/capacity", nil)
	req.Header.Set("Authorization", "Bearer "+rmTok)
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
}
