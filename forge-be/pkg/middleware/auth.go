package middleware

import (
	"net/http"
	"strings"

	"github.com/google/uuid"

	"workspace/internal/auth"
	"workspace/pkg/apperr"
	"workspace/pkg/authctx"
	"workspace/pkg/response"
)

func Authenticate(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if header == "" || !strings.HasPrefix(header, "Bearer ") {
				response.Error(w, apperr.ErrUnauthorized)
				return
			}
			raw := strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
			claims, err := auth.ParseAccess(jwtSecret, raw)
			if err != nil {
				response.Error(w, apperr.ErrUnauthorized)
				return
			}
			uid, err := uuid.Parse(claims.Subject)
			if err != nil {
				response.Error(w, apperr.ErrUnauthorized)
				return
			}
			p := authctx.Principal{
				UserID:      uid,
				RoleCode:    claims.Role,
				Permissions: claims.Perms,
			}
			next.ServeHTTP(w, r.WithContext(authctx.WithUser(r.Context(), p)))
		})
	}
}

func Require(code string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			p, ok := authctx.User(r.Context())
			if !ok {
				response.Error(w, apperr.ErrUnauthorized)
				return
			}
			if !authctx.HasPermission(p, code) {
				response.Error(w, apperr.ErrForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
