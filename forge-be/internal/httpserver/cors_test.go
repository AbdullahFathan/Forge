package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"
)

func TestCORSAllowsLocalFrontend(t *testing.T) {
	r := chi.NewRouter()
	r.Use(corsHandler([]string{"http://localhost:3000", "http://127.0.0.1:3000"}))
	r.Post("/auth/login", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	for _, origin := range []string{"http://localhost:3000", "http://127.0.0.1:3000"} {
		req := httptest.NewRequest(http.MethodOptions, "/auth/login", nil)
		req.Header.Set("Origin", origin)
		req.Header.Set("Access-Control-Request-Method", "POST")
		req.Header.Set("Access-Control-Request-Headers", "content-type,authorization")
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Code, origin)
		require.Equal(t, origin, rec.Header().Get("Access-Control-Allow-Origin"), origin)
		require.Equal(t, "true", rec.Header().Get("Access-Control-Allow-Credentials"), origin)
	}

	req := httptest.NewRequest(http.MethodOptions, "/auth/login", nil)
	req.Header.Set("Origin", "http://evil.example")
	req.Header.Set("Access-Control-Request-Method", "POST")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	require.Empty(t, rec.Header().Get("Access-Control-Allow-Origin"))
}
