package database

import (
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

type Pool struct {
	MaxOpen         int
	MaxIdle         int
	ConnMaxLifetime time.Duration
}

func (p Pool) applyDefaults() Pool {
	if p.MaxOpen <= 0 {
		p.MaxOpen = 25
	}
	if p.MaxIdle <= 0 {
		p.MaxIdle = 5
	}
	if p.ConnMaxLifetime <= 0 {
		p.ConnMaxLifetime = time.Hour
	}
	return p
}

func Connect(dsn string, local bool, log *zap.Logger, pool Pool) (*gorm.DB, error) {
	pool = pool.applyDefaults()
	level := gormlogger.Warn
	if local {
		level = gormlogger.Info
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(level),
	})
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(pool.MaxOpen)
	sqlDB.SetMaxIdleConns(pool.MaxIdle)
	sqlDB.SetConnMaxLifetime(pool.ConnMaxLifetime)
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("database ping: %w", err)
	}
	log.Info("database connected")
	return db, nil
}
