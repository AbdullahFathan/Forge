package auth_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	"workspace/internal/auth"
	"workspace/internal/rbac"
	"workspace/internal/rbac/perm"
	"workspace/internal/user"
	"workspace/pkg/apperr"
)

type fakeUsers struct {
	byEmail map[string]*user.User
	byID    map[uuid.UUID]*user.User
}

func (f *fakeUsers) GetByEmail(email string) (*user.User, error) {
	u, ok := f.byEmail[email]
	if !ok {
		return nil, apperr.ErrNotFound.WithMessage("user not found")
	}
	return u, nil
}

func (f *fakeUsers) GetByID(id uuid.UUID) (*user.User, error) {
	u, ok := f.byID[id]
	if !ok {
		return nil, apperr.ErrNotFound.WithMessage("user not found")
	}
	return u, nil
}

func testUser(t *testing.T, email, password string, active bool, role string) *user.User {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	require.NoError(t, err)
	id := uuid.New()
	u := &user.User{
		ID:           id,
		Name:         "Test",
		Email:        email,
		PasswordHash: string(hash),
		IsActive:     active,
		Role: rbac.Role{
			Code: role,
			Permissions: []rbac.Permission{
				{Code: perm.UserManage},
			},
		},
	}
	return u
}

func TestLoginRefreshLogout(t *testing.T) {
	u := testUser(t, "a@b.com", "password12", true, perm.RoleAdmin)
	users := &fakeUsers{
		byEmail: map[string]*user.User{u.Email: u},
		byID:    map[uuid.UUID]*user.User{u.ID: u},
	}
	store := auth.NewMemoryStore()
	svc := auth.NewService(users, store, "test-secret-at-least-32-characters-long", time.Minute, time.Hour)

	ctx := context.Background()
	tok, err := svc.Login(ctx, "a@b.com", "wrong", "1.1.1.1")
	require.Error(t, err)
	require.Nil(t, tok)
	ae, _ := apperr.As(err)
	require.Equal(t, apperr.ErrInvalidCredentials.Code, ae.Code)

	tok, err = svc.Login(ctx, "a@b.com", "password12", "1.1.1.1")
	require.NoError(t, err)
	require.NotEmpty(t, tok.AccessToken)
	require.NotEmpty(t, tok.Refresh)

	old := tok.Refresh
	tok2, err := svc.Refresh(ctx, old)
	require.NoError(t, err)
	require.NotEqual(t, old, tok2.Refresh)

	_, err = svc.Refresh(ctx, old)
	require.Error(t, err)

	require.NoError(t, svc.Logout(ctx, tok2.Refresh))
	_, err = svc.Refresh(ctx, tok2.Refresh)
	require.Error(t, err)
}

func TestLoginInactive(t *testing.T) {
	u := testUser(t, "x@y.com", "password12", false, perm.RoleMember)
	users := &fakeUsers{
		byEmail: map[string]*user.User{u.Email: u},
		byID:    map[uuid.UUID]*user.User{u.ID: u},
	}
	svc := auth.NewService(users, auth.NewMemoryStore(), "test-secret-at-least-32-characters-long", time.Minute, time.Hour)
	_, err := svc.Login(context.Background(), "x@y.com", "password12", "2.2.2.2")
	ae, _ := apperr.As(err)
	require.Equal(t, apperr.ErrInactiveUser.Code, ae.Code)
}

func TestLoginRateLimit(t *testing.T) {
	u := testUser(t, "a@b.com", "password12", true, perm.RoleAdmin)
	users := &fakeUsers{byEmail: map[string]*user.User{u.Email: u}, byID: map[uuid.UUID]*user.User{u.ID: u}}
	svc := auth.NewService(users, auth.NewMemoryStore(), "test-secret-at-least-32-characters-long", time.Minute, time.Hour)
	ctx := context.Background()
	for i := 0; i < 10; i++ {
		_, err := svc.Login(ctx, "missing@x.com", "nope", "9.9.9.9")
		require.Error(t, err)
	}
	_, err := svc.Login(ctx, "missing@x.com", "nope", "9.9.9.9")
	ae, _ := apperr.As(err)
	require.Equal(t, apperr.ErrRateLimited.Code, ae.Code)
}
