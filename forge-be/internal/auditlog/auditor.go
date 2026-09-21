package auditlog

import (
	"context"

	"github.com/google/uuid"
)

type Record struct {
	ActorID    uuid.UUID
	IP         string
	EntityType string
	EntityID   uuid.UUID
	Action     string
	Before     any
	After      any
	ProjectID  *uuid.UUID
}

type Auditor interface {
	Record(ctx context.Context, rec Record) error
}

type Noop struct{}

func (Noop) Record(context.Context, Record) error {
	return nil
}

func Ptr(id uuid.UUID) *uuid.UUID {
	if id == uuid.Nil {
		return nil
	}
	p := id
	return &p
}
