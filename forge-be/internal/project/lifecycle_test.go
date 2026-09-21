package project

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCanTransition(t *testing.T) {
	require.True(t, CanTransition(StatusDraft, StatusActive))
	require.True(t, CanTransition(StatusActive, StatusCompleted))
	require.True(t, CanTransition(StatusOnHold, StatusArchived))
	require.True(t, CanTransition(StatusCompleted, StatusArchived))
	require.False(t, CanTransition(StatusDraft, StatusArchived))
	require.False(t, CanTransition(StatusActive, StatusDraft))
	require.False(t, CanTransition(StatusArchived, StatusActive))
	require.NoError(t, ValidateTransition(StatusDraft, StatusDraft))
	require.Error(t, ValidateTransition(StatusCompleted, StatusActive))
}
