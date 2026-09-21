package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"workspace/pkg/apperr"
)

type Claims struct {
	Role  string   `json:"role"`
	Perms []string `json:"perms"`
	jwt.RegisteredClaims
}

func SignAccess(secret string, userID uuid.UUID, role string, perms []string, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := Claims{
		Role:  role,
		Perms: perms,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			ID:        uuid.NewString(),
		},
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return tok.SignedString([]byte(secret))
}

func ParseAccess(secret, token string) (*Claims, error) {
	parsed, err := jwt.ParseWithClaims(token, &Claims{}, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, apperr.ErrInvalidToken
		}
		return []byte(secret), nil
	})
	if err != nil || !parsed.Valid {
		return nil, apperr.ErrInvalidToken
	}
	claims, ok := parsed.Claims.(*Claims)
	if !ok {
		return nil, apperr.ErrInvalidToken
	}
	return claims, nil
}
