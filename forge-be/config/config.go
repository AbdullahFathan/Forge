package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

var loadDotenv = true

type Config struct {
	AppEnv        string
	HTTPPort      string
	DatabaseDSN   string
	RedisAddr     string
	RedisPassword string

	JWTSecret     string
	JWTAccessTTL  time.Duration
	JWTRefreshTTL time.Duration

	CookieSecure bool
	CORSOrigins  []string

	BootstrapAdminEmail    string
	BootstrapAdminPassword string

	RustFSEndpoint  string
	RustFSAccessKey string
	RustFSSecretKey string
	RustFSBucket    string
	RustFSUseSSL    bool

	DBMaxOpenConns    int
	DBMaxIdleConns    int
	DBConnMaxLifetime time.Duration

	SMTPHost     string
	SMTPPort     string
	SMTPUser     string
	SMTPPassword string
	SMTPFrom     string
	SMTPStartTLS bool
}

func Load() (*Config, error) {
	if loadDotenv {
		_ = godotenv.Load()
		_ = godotenv.Load("../.env")
	}

	cfg := &Config{
		AppEnv:                 envOr("APP_ENV", "local"),
		HTTPPort:               envOr("HTTP_PORT", "8080"),
		DatabaseDSN:            os.Getenv("DATABASE_DSN"),
		RedisAddr:              os.Getenv("REDIS_ADDR"),
		RedisPassword:          os.Getenv("REDIS_PASSWORD"),
		JWTSecret:              os.Getenv("JWT_SECRET"),
		CookieSecure:           envBool("COOKIE_SECURE", false),
		BootstrapAdminEmail:    os.Getenv("BOOTSTRAP_ADMIN_EMAIL"),
		BootstrapAdminPassword: os.Getenv("BOOTSTRAP_ADMIN_PASSWORD"),
		RustFSEndpoint:         envOr("RUSTFS_ENDPOINT", "http://rustfs:9000"),
		RustFSAccessKey:        os.Getenv("RUSTFS_ACCESS_KEY"),
		RustFSSecretKey:        os.Getenv("RUSTFS_SECRET_KEY"),
		RustFSBucket:           envOr("RUSTFS_BUCKET", "workspace"),
		RustFSUseSSL:           envBool("RUSTFS_USE_SSL", false),
		DBMaxOpenConns:         envInt("DB_MAX_OPEN_CONNS", 25),
		DBMaxIdleConns:         envInt("DB_MAX_IDLE_CONNS", 5),
		SMTPHost:               os.Getenv("SMTP_HOST"),
		SMTPPort:               envOr("SMTP_PORT", "587"),
		SMTPUser:               os.Getenv("SMTP_USER"),
		SMTPPassword:           os.Getenv("SMTP_PASSWORD"),
		SMTPFrom:               os.Getenv("SMTP_FROM"),
		SMTPStartTLS:           envBool("SMTP_STARTTLS", true),
	}

	var err error
	cfg.JWTAccessTTL, err = envDuration("JWT_ACCESS_TTL", 15*time.Minute)
	if err != nil {
		return nil, err
	}
	cfg.JWTRefreshTTL, err = envDuration("JWT_REFRESH_TTL", 7*24*time.Hour)
	if err != nil {
		return nil, err
	}
	cfg.DBConnMaxLifetime, err = envDuration("DB_CONN_MAX_LIFETIME", time.Hour)
	if err != nil {
		return nil, err
	}

	origins := envOr("CORS_ORIGINS", "")
	for _, o := range strings.Split(origins, ",") {
		o = strings.TrimSpace(o)
		if o != "" {
			cfg.CORSOrigins = append(cfg.CORSOrigins, o)
		}
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (c *Config) IsLocal() bool {
	return c.AppEnv == "local" || c.AppEnv == "development" || c.AppEnv == "dev"
}

func (c *Config) validate() error {
	required := map[string]string{
		"DATABASE_DSN":             c.DatabaseDSN,
		"REDIS_ADDR":               c.RedisAddr,
		"JWT_SECRET":               c.JWTSecret,
		"BOOTSTRAP_ADMIN_EMAIL":    c.BootstrapAdminEmail,
		"BOOTSTRAP_ADMIN_PASSWORD": c.BootstrapAdminPassword,
	}
	var missing []string
	for k, v := range required {
		if strings.TrimSpace(v) == "" {
			missing = append(missing, k)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing required env: %s", strings.Join(missing, ", "))
	}
	if len(c.JWTSecret) < 32 {
		return fmt.Errorf("JWT_SECRET must be at least 32 characters")
	}
	if len(c.BootstrapAdminPassword) < 8 {
		return fmt.Errorf("BOOTSTRAP_ADMIN_PASSWORD must be at least 8 characters")
	}
	return nil
}

func envInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envBool(key string, fallback bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return fallback
	}
	return b
}

func envDuration(key string, fallback time.Duration) (time.Duration, error) {
	v := os.Getenv(key)
	if v == "" {
		return fallback, nil
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return 0, fmt.Errorf("invalid %s: %w", key, err)
	}
	return d, nil
}
