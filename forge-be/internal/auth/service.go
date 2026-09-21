package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"workspace/internal/rbac/perm"
	"workspace/internal/user"
	"workspace/pkg/apperr"
)

type UserReader interface {
	GetByEmail(email string) (*user.User, error)
	GetByID(id uuid.UUID) (*user.User, error)
}

type Service struct {
	users      UserReader
	tokens     TokenStore
	jwtSecret  string
	accessTTL  time.Duration
	refreshTTL time.Duration
}

type Tokens struct {
	AccessToken string
	ExpiresIn   int
	Refresh     string
	FamilyID    string
	User        *user.User
}

func NewService(users UserReader, tokens TokenStore, jwtSecret string, accessTTL, refreshTTL time.Duration) *Service {
	return &Service{
		users:      users,
		tokens:     tokens,
		jwtSecret:  jwtSecret,
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
	}
}

func (s *Service) Login(ctx context.Context, email, password, ip string) (*Tokens, error) {
	ok, err := s.tokens.AllowLogin(ctx, ip, 10, time.Minute)
	if err != nil {
		return nil, apperr.ErrInternal
	}
	if !ok {
		return nil, apperr.ErrRateLimited
	}

	u, err := s.users.GetByEmail(email)
	if err != nil {
		if errors.Is(err, apperr.ErrNotFound) || apperrIsNotFound(err) {
			return nil, apperr.ErrInvalidCredentials
		}
		return nil, err
	}
	if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)) != nil {
		return nil, apperr.ErrInvalidCredentials
	}
	if !u.IsActive {
		return nil, apperr.ErrInactiveUser
	}
	return s.issue(ctx, u, uuid.NewString())
}

func (s *Service) Refresh(ctx context.Context, refreshToken string) (*Tokens, error) {
	userID, familyID, err := s.tokens.GetRefresh(ctx, refreshToken)
	if err != nil {
		return nil, apperr.ErrInvalidToken
	}
	valid, err := s.tokens.FamilyValid(ctx, familyID)
	if err != nil {
		return nil, apperr.ErrInternal
	}
	if !valid {
		return nil, apperr.ErrInvalidToken
	}
	_ = s.tokens.DeleteRefresh(ctx, refreshToken)

	id, err := uuid.Parse(userID)
	if err != nil {
		return nil, apperr.ErrInvalidToken
	}
	u, err := s.users.GetByID(id)
	if err != nil {
		return nil, apperr.ErrInvalidToken
	}
	if !u.IsActive {
		_ = s.tokens.InvalidateFamily(ctx, familyID)
		return nil, apperr.ErrInactiveUser
	}
	return s.issue(ctx, u, familyID)
}

func (s *Service) Logout(ctx context.Context, refreshToken string) error {
	if refreshToken == "" {
		return nil
	}
	_, familyID, err := s.tokens.GetRefresh(ctx, refreshToken)
	if err != nil {
		return nil
	}
	_ = s.tokens.DeleteRefresh(ctx, refreshToken)
	return s.tokens.InvalidateFamily(ctx, familyID)
}

func (s *Service) issue(ctx context.Context, u *user.User, familyID string) (*Tokens, error) {
	perms := make([]string, 0, len(u.Role.Permissions))
	for _, p := range u.Role.Permissions {
		perms = append(perms, p.Code)
	}
	if u.Role.Code == perm.RoleSuperAdmin && len(perms) == 0 {
		perms = perm.All()
	}
	access, err := SignAccess(s.jwtSecret, u.ID, u.Role.Code, perms, s.accessTTL)
	if err != nil {
		return nil, err
	}
	refresh, err := randomToken()
	if err != nil {
		return nil, err
	}
	if err := s.tokens.SaveRefresh(ctx, refresh, u.ID.String(), familyID, s.refreshTTL); err != nil {
		return nil, apperr.ErrInternal
	}
	return &Tokens{
		AccessToken: access,
		ExpiresIn:   int(s.accessTTL.Seconds()),
		Refresh:     refresh,
		FamilyID:    familyID,
		User:        u,
	}, nil
}

func randomToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func apperrIsNotFound(err error) bool {
	ae, ok := apperr.As(err)
	return ok && ae.Code == apperr.ErrNotFound.Code
}
