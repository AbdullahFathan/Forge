package rbac

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"

	"workspace/internal/rbac/perm"
)

func roleRouter(h *Handler) http.Handler {
	r := chi.NewRouter()
	r.Get("/permissions", h.ListPermissions)
	r.Get("/roles", h.List)
	r.Post("/roles", h.Create)
	r.Get("/roles/{id}", h.Get)
	r.Patch("/roles/{id}", h.Patch)
	r.Delete("/roles/{id}", h.Delete)
	return r
}

func TestRoleHandlerCRUD(t *testing.T) {
	h := NewHandler(NewService(newMem()))
	rt := roleRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/permissions", nil)
	rec := httptest.NewRecorder()
	rt.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	body, _ := json.Marshal(map[string]any{
		"name": "Ops Lead", "permissionCodes": []string{perm.TaskManage},
	})
	req = httptest.NewRequest(http.MethodPost, "/roles", bytes.NewReader(body))
	rec = httptest.NewRecorder()
	rt.ServeHTTP(rec, req)
	require.Equal(t, http.StatusCreated, rec.Code)

	var env struct {
		Data Public `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &env))
	id := env.Data.ID.String()

	req = httptest.NewRequest(http.MethodGet, "/roles/"+id, nil)
	rec = httptest.NewRecorder()
	rt.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	patch, _ := json.Marshal(map[string]any{"name": "Ops"})
	req = httptest.NewRequest(http.MethodPatch, "/roles/"+id, bytes.NewReader(patch))
	rec = httptest.NewRecorder()
	rt.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	req = httptest.NewRequest(http.MethodGet, "/roles", nil)
	rec = httptest.NewRecorder()
	rt.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	req = httptest.NewRequest(http.MethodDelete, "/roles/"+id, nil)
	rec = httptest.NewRecorder()
	rt.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	bad, _ := json.Marshal(map[string]any{"name": "X", "permissionCodes": []string{"nope"}})
	req = httptest.NewRequest(http.MethodPost, "/roles", bytes.NewReader(bad))
	rec = httptest.NewRecorder()
	rt.ServeHTTP(rec, req)
	require.Equal(t, http.StatusBadRequest, rec.Code)
}
