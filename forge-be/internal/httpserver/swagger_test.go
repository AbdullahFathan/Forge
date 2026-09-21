package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSwaggerAndHealthHandlers(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/swagger/index.html", nil)
	rec := httptest.NewRecorder()
	SwaggerUI().ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
	require.Contains(t, rec.Body.String(), "swagger-ui")

	req = httptest.NewRequest(http.MethodGet, "/swagger/openapi.yaml", nil)
	rec = httptest.NewRecorder()
	SwaggerSpec().ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
	require.Contains(t, rec.Body.String(), "/auth/login")
	require.Contains(t, rec.Body.String(), "/users")
	require.Contains(t, rec.Body.String(), "/departments")
	require.Contains(t, rec.Body.String(), "/projects")
	require.Contains(t, rec.Body.String(), "/tasks/{id}")
	require.Contains(t, rec.Body.String(), "/resources/allocations")
	require.Contains(t, rec.Body.String(), "/resources/capacity")
	require.Contains(t, rec.Body.String(), "/me/workload")
}
