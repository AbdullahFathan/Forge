package task

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestStatusMachineAndCycle(t *testing.T) {
	require.True(t, CanTransition(StatusBacklog, StatusTodo))
	require.True(t, CanTransition(StatusInProgress, StatusDone))
	require.False(t, CanTransition(StatusBacklog, StatusDone))
	require.True(t, IsStartStatus(StatusTodo))
	require.False(t, IsStartStatus(StatusDone))

	a, b, c := uuid.New(), uuid.New(), uuid.New()
	edges := map[uuid.UUID][]uuid.UUID{a: {b}, b: {c}}
	require.True(t, WouldCycle(edges, c, a))
	require.False(t, WouldCycle(edges, a, uuid.New()))
}

func TestCompletionMath(t *testing.T) {
	done, total := int64(3), int64(10)
	pct := float64(done) / float64(total) * 100
	require.Equal(t, 30.0, pct)
}
