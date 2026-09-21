package dashboard

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"workspace/internal/rbac/perm"
	"workspace/pkg/apperr"
	"workspace/pkg/authctx"
)

func TestExecutiveForbiddenForMember(t *testing.T) {
	s := &Service{clock: realClock{}}
	_, err := s.Executive(authctx.Principal{UserID: uuid.New(), RoleCode: perm.RoleMember})
	require.Equal(t, apperr.ErrForbidden.Code, apperrMust(err))
	_, err = s.ProjectManager(authctx.Principal{UserID: uuid.New(), RoleCode: perm.RoleMember})
	require.Equal(t, apperr.ErrForbidden.Code, apperrMust(err))
}

func apperrMust(err error) string {
	ae, ok := apperr.As(err)
	if !ok {
		return ""
	}
	return ae.Code
}
