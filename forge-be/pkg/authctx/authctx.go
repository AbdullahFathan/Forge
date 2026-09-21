package authctx

import (
	"context"

	"github.com/google/uuid"
)

type ctxKey int

const (
	userKey ctxKey = iota
	requestIDKey
)

type Principal struct {
	UserID      uuid.UUID
	RoleCode    string
	Permissions []string
}

func WithUser(ctx context.Context, p Principal) context.Context {
	return context.WithValue(ctx, userKey, p)
}

func User(ctx context.Context) (Principal, bool) {
	p, ok := ctx.Value(userKey).(Principal)
	return p, ok
}

func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey, id)
}

func RequestID(ctx context.Context) string {
	id, _ := ctx.Value(requestIDKey).(string)
	return id
}

func HasPermission(p Principal, code string) bool {
	for _, c := range p.Permissions {
		if c == code {
			return true
		}
	}
	return false
}
