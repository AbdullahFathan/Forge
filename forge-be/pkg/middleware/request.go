package middleware

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"workspace/pkg/authctx"
)

func RequestLogger(log *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			reqID := r.Header.Get("X-Request-ID")
			if reqID == "" {
				reqID = uuid.NewString()
			}
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			ww.Header().Set("X-Request-ID", reqID)
			ctx := authctx.WithRequestID(r.Context(), reqID)
			next.ServeHTTP(ww, r.WithContext(ctx))
			log.Info("request",
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.Int("status", ww.Status()),
				zap.Duration("duration", time.Since(start)),
				zap.String("request_id", reqID),
			)
		})
	}
}
