package user

import (
	"testing"

	"github.com/stretchr/testify/require"

	"workspace/internal/rbac/perm"
	"workspace/pkg/apperr"
)

func TestGuardLastSuperAdmin(t *testing.T) {
	tests := []struct {
		name    string
		role    string
		remove  bool
		count   int64
		wantErr bool
	}{
		{"member delete ok", perm.RoleMember, true, 1, false},
		{"sa not removing ok", perm.RoleSuperAdmin, false, 1, false},
		{"last sa blocked", perm.RoleSuperAdmin, true, 1, true},
		{"second sa ok", perm.RoleSuperAdmin, true, 2, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := GuardLastSuperAdmin(tc.role, tc.remove, tc.count)
			if tc.wantErr {
				ae, ok := apperr.As(err)
				require.True(t, ok)
				require.Equal(t, apperr.ErrLastSuperAdmin.Code, ae.Code)
				return
			}
			require.NoError(t, err)
		})
	}
}
