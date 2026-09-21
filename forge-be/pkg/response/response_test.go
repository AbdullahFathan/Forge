package response

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"workspace/pkg/apperr"
)

func TestJSONAndError(t *testing.T) {
	rec := httptest.NewRecorder()
	JSON(rec, http.StatusOK, map[string]string{"ok": "1"})
	var env Envelope
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &env))
	require.True(t, env.Success)

	rec = httptest.NewRecorder()
	Error(rec, apperr.ErrForbidden)
	require.Equal(t, http.StatusForbidden, rec.Code)
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &env))
	require.False(t, env.Success)
	require.Equal(t, apperr.ErrForbidden.Code, env.Error.Code)
}

func TestDecodeUnknownField(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"x":1}`))
	var dst struct {
		Y int `json:"y"`
	}
	err := Decode(req, &dst)
	require.Error(t, err)
}
