package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEvaluateReady(t *testing.T) {
	code, payload := EvaluateReady(nil, nil)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, "ready", payload["status"])

	code, payload = EvaluateReady(errors.New("db"), nil)
	require.Equal(t, http.StatusServiceUnavailable, code)
	require.Equal(t, "not_ready", payload["status"])
	checks := payload["checks"].(map[string]string)
	require.Equal(t, "fail", checks["db"])
	require.Equal(t, "ok", checks["redis"])

	code, payload = EvaluateReady(nil, errors.New("redis"))
	require.Equal(t, http.StatusServiceUnavailable, code)
	checks = payload["checks"].(map[string]string)
	require.Equal(t, "fail", checks["redis"])
}

func TestReadyHandlerNilDeps(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/ready", nil)
	rec := httptest.NewRecorder()
	Ready(nil, nil).ServeHTTP(rec, req)
	require.Equal(t, http.StatusServiceUnavailable, rec.Code)
	var env struct {
		Success bool `json:"success"`
		Data    struct {
			Status string            `json:"status"`
			Checks map[string]string `json:"checks"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &env))
	require.Equal(t, "not_ready", env.Data.Status)
	require.Equal(t, "fail", env.Data.Checks["db"])
	require.Equal(t, "fail", env.Data.Checks["redis"])
}
