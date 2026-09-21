package notification

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type memPref struct{ on bool }

func (m memPref) EmailNotificationsEnabled(uuid.UUID) (bool, error) { return m.on, nil }

type countSender struct{ n int }

func (c *countSender) Send(context.Context, uuid.UUID, Event) error {
	c.n++
	return nil
}

func TestNopSender(t *testing.T) {
	require.NoError(t, NopSender{}.Send(context.Background(), uuid.New(), Event{}))
}

func TestDateUTC(t *testing.T) {
	in, err := time.Parse(time.RFC3339, "2026-03-15T18:00:00Z")
	require.NoError(t, err)
	d := DateUTC(in)
	require.Equal(t, 0, d.Hour())
	require.Equal(t, 15, d.Day())
}

func TestEmailBranchIncrements(t *testing.T) {
	cs := &countSender{}
	prefOn := memPref{on: true}
	ok, err := prefOn.EmailNotificationsEnabled(uuid.New())
	require.NoError(t, err)
	require.True(t, ok)
	_ = cs.Send(context.Background(), uuid.New(), Event{Type: TypeTaskAssigned})
	require.Equal(t, 1, cs.n)

	prefOff := memPref{on: false}
	ok, err = prefOff.EmailNotificationsEnabled(uuid.New())
	require.NoError(t, err)
	require.False(t, ok)
}
