package apperr

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAsAndWrap(t *testing.T) {
	_, ok := As(errors.New("plain"))
	require.False(t, ok)
	ae, ok := As(ErrForbidden.WithMessage("no"))
	require.True(t, ok)
	require.Equal(t, "no", ae.Message)
	require.Equal(t, http.StatusForbidden, ae.Status)

	require.Nil(t, WrapInternal(nil))
	wrapped := WrapInternal(errors.New("db"))
	inner, ok := As(ErrInternal)
	require.True(t, ok)
	require.Error(t, wrapped)
	require.Equal(t, "INTERNAL_ERROR", inner.Code)
}
