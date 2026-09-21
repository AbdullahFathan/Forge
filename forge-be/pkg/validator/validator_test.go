package validator

import (
	"testing"

	"github.com/stretchr/testify/require"

	"workspace/pkg/apperr"
)

func TestStruct(t *testing.T) {
	type req struct {
		Name string `validate:"required,min=2"`
	}
	err := Struct(req{Name: "ab"})
	require.NoError(t, err)
	err = Struct(req{})
	require.Error(t, err)
	ae, ok := apperr.As(err)
	require.True(t, ok)
	require.Equal(t, apperr.ErrValidation.Code, ae.Code)
}
