package migrate_test

import (
	"os"
	"testing"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"workspace/internal/migrate"
)

func TestUpAgainstComposePostgres(t *testing.T) {
	_ = godotenv.Load("../../.env")
	dsn := os.Getenv("DATABASE_DSN")
	if dsn == "" {
		dsn = "postgres://workspace:workspace@localhost:5432/workspace?sslmode=disable"
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: gormlogger.Default.LogMode(gormlogger.Silent)})
	if err != nil {
		t.Skip("postgres not available: ", err)
	}
	sqlDB, err := db.DB()
	require.NoError(t, err)
	if err := sqlDB.Ping(); err != nil {
		t.Skip("postgres ping failed: ", err)
	}
	require.NoError(t, migrate.Up(db))
	require.NoError(t, migrate.Up(db))

	var n int64
	require.NoError(t, db.Raw(`SELECT COUNT(*) FROM schema_migrations`).Scan(&n).Error)
	require.GreaterOrEqual(t, n, int64(3))
	require.NoError(t, db.Raw(`SELECT COUNT(*) FROM information_schema.tables WHERE table_name = 'resource_allocations'`).Scan(&n).Error)
	require.Equal(t, int64(1), n)
}
