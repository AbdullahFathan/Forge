package user_test

import (
	"bytes"
	"encoding/json"
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

func TestUserRoutesForbiddenForMember(t *testing.T) {
	secret := "test-secret-at-least-32-characters-long"
	tok, err := auth.SignAccess(secret, uuid.New(), perm.RoleMember, []string{perm.TaskManage}, time.Hour)
	require.NoError(t, err)

	r := chi.NewRouter()
	r.Group(func(r chi.Router) {
		r.Use(middleware.Authenticate(secret))
		r.With(middleware.Require(perm.UserManage)).Get("/users", func(w http.ResponseWriter, r *http.Request) {
			response.JSON(w, http.StatusOK, map[string]string{"ok": "1"})
		})
		r.With(middleware.Require(perm.DepartmentManage)).Post("/departments", func(w http.ResponseWriter, r *http.Request) {
			response.JSON(w, http.StatusCreated, map[string]string{"ok": "1"})
		})
	})

	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	require.Equal(t, http.StatusForbidden, rec.Code)

	body, _ := json.Marshal(map[string]string{"name": "X"})
	req = httptest.NewRequest(http.MethodPost, "/departments", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+tok)
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	require.Equal(t, http.StatusForbidden, rec.Code)
}
