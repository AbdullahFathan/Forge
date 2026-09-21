package rbac_test

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

func TestRoleRoutesForbiddenWithoutRoleManage(t *testing.T) {
	secret := "test-secret-at-least-32-characters-long"
	tok, err := auth.SignAccess(secret, uuid.New(), perm.RoleMember, []string{perm.TaskManage}, time.Hour)
	require.NoError(t, err)

	r := chi.NewRouter()
	r.Group(func(r chi.Router) {
		r.Use(middleware.Authenticate(secret))
		r.With(middleware.Require(perm.RoleManage)).Get("/roles", func(w http.ResponseWriter, r *http.Request) {
			response.JSON(w, http.StatusOK, []any{})
		})
	})

	req := httptest.NewRequest(http.MethodGet, "/roles", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	require.Equal(t, http.StatusForbidden, rec.Code)
}
