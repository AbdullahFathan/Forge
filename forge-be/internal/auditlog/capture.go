package auditlog

import "context"

// Capture is an in-memory Auditor for tests.
type Capture struct {
	Items []Record
}

func (c *Capture) Record(_ context.Context, rec Record) error {
	c.Items = append(c.Items, rec)
	return nil
}
