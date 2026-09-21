package project_test

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

func TestProjectRoutesRBAC(t *testing.T) {
	secret := "test-secret-at-least-32-characters-long"
	memberTok, err := auth.SignAccess(secret, uuid.New(), perm.RoleMember, []string{perm.TaskManage}, time.Hour)
	require.NoError(t, err)
	pmTok, err := auth.SignAccess(secret, uuid.New(), perm.RoleProjectManager, []string{perm.ProjectCreate, perm.ProjectDelete, perm.TaskManage}, time.Hour)
	require.NoError(t, err)

	r := chi.NewRouter()
	r.Group(func(r chi.Router) {
		r.Use(middleware.Authenticate(secret))
		r.Get("/projects", func(w http.ResponseWriter, r *http.Request) {
			response.JSON(w, http.StatusOK, map[string]string{"ok": "1"})
		})
		r.With(middleware.Require(perm.ProjectCreate)).Post("/projects", func(w http.ResponseWriter, r *http.Request) {
			response.JSON(w, http.StatusCreated, map[string]string{"ok": "1"})
		})
		r.With(middleware.Require(perm.ProjectDelete)).Delete("/projects/{id}", func(w http.ResponseWriter, r *http.Request) {
			response.JSON(w, http.StatusOK, map[string]string{"ok": "1"})
		})
	})

	req := httptest.NewRequest(http.MethodGet, "/projects", nil)
	req.Header.Set("Authorization", "Bearer "+memberTok)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	req = httptest.NewRequest(http.MethodPost, "/projects", nil)
	req.Header.Set("Authorization", "Bearer "+memberTok)
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	require.Equal(t, http.StatusForbidden, rec.Code)

	req = httptest.NewRequest(http.MethodPost, "/projects", nil)
	req.Header.Set("Authorization", "Bearer "+pmTok)
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	require.Equal(t, http.StatusCreated, rec.Code)
}
