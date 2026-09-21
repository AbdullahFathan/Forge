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
}

type Auditor interface {
	Record(ctx context.Context, rec Record) error
}

type Noop struct{}

func (Noop) Record(context.Context, Record) error {
	return nil
}
