package notification

import "context"

type Capture struct {
	Events []Event
}

func (c *Capture) Emit(_ context.Context, ev Event) error {
	c.Events = append(c.Events, ev)
	return nil
}

type NopEmitter struct{}

func (NopEmitter) Emit(context.Context, Event) error { return nil }
