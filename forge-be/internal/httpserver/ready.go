package server

import (
	"net/http"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"workspace/pkg/response"
)

func EvaluateReady(dbErr, redisErr error) (status int, payload map[string]any) {
	checks := map[string]string{
		"db":    "ok",
		"redis": "ok",
	}
	ok := true
	if dbErr != nil {
		checks["db"] = "fail"
		ok = false
	}
	if redisErr != nil {
		checks["redis"] = "fail"
		ok = false
	}
	if ok {
		return http.StatusOK, map[string]any{"status": "ready", "checks": checks}
	}
	return http.StatusServiceUnavailable, map[string]any{"status": "not_ready", "checks": checks}
}

func Ready(db *gorm.DB, rdb *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var dbErr error
		if db == nil {
			dbErr = errDBUnavailable
		} else {
			sqlDB, err := db.DB()
			if err != nil {
				dbErr = err
			} else {
				dbErr = sqlDB.Ping()
			}
		}
		var redisErr error
		if rdb == nil {
			redisErr = errRedisUnavailable
		} else {
			redisErr = rdb.Ping(r.Context()).Err()
		}
		code, payload := EvaluateReady(dbErr, redisErr)
		response.JSON(w, code, payload)
	}
}

var (
	errDBUnavailable    = errString("database unavailable")
	errRedisUnavailable = errString("redis unavailable")
)

type errString string

func (e errString) Error() string { return string(e) }
