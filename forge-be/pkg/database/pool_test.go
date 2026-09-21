package database

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestPoolDefaults(t *testing.T) {
	p := Pool{}.applyDefaults()
	require.Equal(t, 25, p.MaxOpen)
	require.Equal(t, 5, p.MaxIdle)
	require.Equal(t, time.Hour, p.ConnMaxLifetime)
}
