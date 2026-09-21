package auth

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"workspace/pkg/apperr"
)

func TestLoginValidation(t *testing.T) {
	h := NewHandler(nil, false, time.Hour)
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString(`{}`))
	rec := httptest.NewRecorder()
	h.Login(rec, req)
	require.Equal(t, http.StatusBadRequest, rec.Code)

	var body map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	errObj := body["error"].(map[string]any)
	require.Equal(t, apperr.ErrValidation.Code, errObj["code"])
}
