package storage

import (
	"bytes"
	"context"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Config struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
	UseSSL    bool
}

type Client interface {
	Put(ctx context.Context, key string, body []byte, contentType string) error
	PresignGet(ctx context.Context, key string, ttl time.Duration) (string, error)
	EnsureBucket(ctx context.Context) error
}

type Nop struct{}

func (Nop) Put(context.Context, string, []byte, string) error { return nil }
func (Nop) PresignGet(context.Context, string, time.Duration) (string, error) {
	return "", nil
}
func (Nop) EnsureBucket(context.Context) error { return nil }

type Minio struct {
	cli    *minio.Client
	bucket string
}

func New(cfg Config) (Client, error) {
	if strings.TrimSpace(cfg.Endpoint) == "" || strings.TrimSpace(cfg.AccessKey) == "" {
		return Nop{}, nil
	}
	endpoint := strings.TrimPrefix(strings.TrimPrefix(cfg.Endpoint, "https://"), "http://")
	cli, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, err
	}
	return &Minio{cli: cli, bucket: cfg.Bucket}, nil
}

func (m *Minio) EnsureBucket(ctx context.Context) error {
	ok, err := m.cli.BucketExists(ctx, m.bucket)
	if err != nil {
		return err
	}
	if ok {
		return nil
	}
	return m.cli.MakeBucket(ctx, m.bucket, minio.MakeBucketOptions{})
}

func (m *Minio) Put(ctx context.Context, key string, body []byte, contentType string) error {
	_, err := m.cli.PutObject(ctx, m.bucket, key, bytes.NewReader(body), int64(len(body)), minio.PutObjectOptions{ContentType: contentType})
	return err
}

func (m *Minio) PresignGet(ctx context.Context, key string, ttl time.Duration) (string, error) {
	u, err := m.cli.PresignedGetObject(ctx, m.bucket, key, ttl, nil)
	if err != nil {
		return "", err
	}
	return u.String(), nil
}
