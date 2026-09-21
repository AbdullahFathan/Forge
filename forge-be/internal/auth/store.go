package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type refreshRecord struct {
	UserID   string `json:"userId"`
	FamilyID string `json:"familyId"`
}

type TokenStore interface {
	SaveRefresh(ctx context.Context, token, userID, familyID string, ttl time.Duration) error
	GetRefresh(ctx context.Context, token string) (userID, familyID string, err error)
	DeleteRefresh(ctx context.Context, token string) error
	InvalidateFamily(ctx context.Context, familyID string) error
	FamilyValid(ctx context.Context, familyID string) (bool, error)
	AllowLogin(ctx context.Context, ip string, limit int, window time.Duration) (bool, error)
}

type RedisStore struct {
	rdb *redis.Client
}

func NewRedisStore(rdb *redis.Client) *RedisStore {
	return &RedisStore{rdb: rdb}
}

func refreshKey(token string) string { return "auth:refresh:" + token }
func familyKey(id string) string     { return "auth:family:" + id }
func loginKey(ip string) string      { return "auth:login:" + ip }

func (s *RedisStore) SaveRefresh(ctx context.Context, token, userID, familyID string, ttl time.Duration) error {
	payload, err := json.Marshal(refreshRecord{UserID: userID, FamilyID: familyID})
	if err != nil {
		return err
	}
	pipe := s.rdb.TxPipeline()
	pipe.Set(ctx, refreshKey(token), payload, ttl)
	pipe.Set(ctx, familyKey(familyID), "1", ttl)
	_, err = pipe.Exec(ctx)
	return err
}

func (s *RedisStore) GetRefresh(ctx context.Context, token string) (string, string, error) {
	raw, err := s.rdb.Get(ctx, refreshKey(token)).Bytes()
	if err == redis.Nil {
		return "", "", fmt.Errorf("not found")
	}
	if err != nil {
		return "", "", err
	}
	var rec refreshRecord
	if err := json.Unmarshal(raw, &rec); err != nil {
		return "", "", err
	}
	return rec.UserID, rec.FamilyID, nil
}

func (s *RedisStore) DeleteRefresh(ctx context.Context, token string) error {
	return s.rdb.Del(ctx, refreshKey(token)).Err()
}

func (s *RedisStore) InvalidateFamily(ctx context.Context, familyID string) error {
	return s.rdb.Del(ctx, familyKey(familyID)).Err()
}

func (s *RedisStore) FamilyValid(ctx context.Context, familyID string) (bool, error) {
	n, err := s.rdb.Exists(ctx, familyKey(familyID)).Result()
	if err != nil {
		return false, err
	}
	return n == 1, nil
}

func (s *RedisStore) AllowLogin(ctx context.Context, ip string, limit int, window time.Duration) (bool, error) {
	key := loginKey(ip)
	n, err := s.rdb.Incr(ctx, key).Result()
	if err != nil {
		return false, err
	}
	if n == 1 {
		if err := s.rdb.Expire(ctx, key, window).Err(); err != nil {
			return false, err
		}
	}
	return n <= int64(limit), nil
}
