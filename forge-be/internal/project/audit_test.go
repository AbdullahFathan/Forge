package project

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"workspace/internal/auditlog"
	"workspace/internal/rbac/perm"
	"workspace/pkg/authctx"
)

type failStore struct{ *fakeStore }

func (f failStore) Create(*Project) error { return errors.New("boom") }

func TestFailedCreateDoesNotAudit(t *testing.T) {
	owner := uuid.New()
	cap := &auditlog.Capture{}
	store := failStore{fakeStore: newFakeStore()}
	svc := NewService(store, fakeUsers{ids: map[uuid.UUID]bool{owner: true}}, fakeDepts{}, fakeStats{}, cap)
	pm := authctx.Principal{UserID: owner, RoleCode: perm.RoleProjectManager, Permissions: []string{perm.ProjectCreate}}
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	_, err := svc.Create(context.Background(), pm, "1.1.1.1", CreateInput{
		Name: "X", OwnerID: owner, StartDate: start, TargetEndDate: end,
	})
	require.Error(t, err)
	require.Empty(t, cap.Items)
}
