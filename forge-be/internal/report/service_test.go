package report

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"

	"workspace/internal/rbac/perm"
	"workspace/pkg/apperr"
	"workspace/pkg/authctx"
)

func TestWriteCSVAndPDF(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, WriteCSV(&buf, []string{"a", "b"}, [][]string{{"1", "2"}}))
	require.Contains(t, buf.String(), "a,b")
	pdf, err := WritePDF("t", []string{"a"}, [][]string{{"1"}})
	require.NoError(t, err)
	require.Greater(t, len(pdf), 100)
}

func TestProjectStatusForbiddenWithoutExport(t *testing.T) {
	s := &Service{}
	_, err := s.ProjectStatus(authctx.Principal{RoleCode: perm.RoleMember, Permissions: []string{perm.TaskManage}}, Filter{})
	require.Equal(t, apperr.ErrForbidden.Code, mustCode(err))
}

func mustCode(err error) string {
	ae, ok := apperr.As(err)
	if !ok {
		return ""
	}
	return ae.Code
}
