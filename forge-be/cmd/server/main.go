package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"workspace/config"
	"workspace/internal/httpserver"
	"workspace/internal/migrate"
	"workspace/internal/seed"
	"workspace/pkg/database"
	"workspace/pkg/logger"
	"workspace/pkg/storage"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}
	log, err := logger.New(cfg.IsLocal())
	if err != nil {
		panic(err)
	}
	defer log.Sync()

	db, err := database.Connect(cfg.DatabaseDSN, cfg.IsLocal(), log, database.Pool{
		MaxOpen:         cfg.DBMaxOpenConns,
		MaxIdle:         cfg.DBMaxIdleConns,
		ConnMaxLifetime: cfg.DBConnMaxLifetime,
	})
	if err != nil {
		log.Fatal("database", zap.Error(err))
	}
	if err := migrate.Up(db); err != nil {
		log.Fatal("migrate", zap.Error(err))
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatal("redis ping", zap.Error(err))
	}

	if err := seed.Run(db, cfg.BootstrapAdminEmail, cfg.BootstrapAdminPassword); err != nil {
		log.Fatal("seed", zap.Error(err))
	}

	store, err := storage.New(storage.Config{
		Endpoint:  cfg.RustFSEndpoint,
		AccessKey: cfg.RustFSAccessKey,
		SecretKey: cfg.RustFSSecretKey,
		Bucket:    cfg.RustFSBucket,
		UseSSL:    cfg.RustFSUseSSL,
	})
	if err != nil {
		log.Warn("storage client", zap.Error(err))
		store = storage.Nop{}
	}
	ctxBucket, cancelBucket := context.WithTimeout(context.Background(), 5*time.Second)
	if err := store.EnsureBucket(ctxBucket); err != nil {
		log.Warn("rustfs bucket", zap.Error(err))
	}
	cancelBucket()

	app := server.NewRouter(cfg, log, db, rdb, store)
	httpServer := &http.Server{
		Addr:              ":" + cfg.HTTPPort,
		Handler:           app.Handler,
		ReadHeaderTimeout: 10 * time.Second,
	}

	jobCtx, jobCancel := context.WithCancel(context.Background())
	go app.Scheduler.Run(jobCtx, 15*time.Minute)

	go func() {
		log.Info("http listening", zap.String("addr", httpServer.Addr))
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("http", zap.Error(err))
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	log.Info("shutdown started")
	jobCancel()
	shCtx, shCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shCancel()
	if err := httpServer.Shutdown(shCtx); err != nil {
		log.Error("http shutdown", zap.Error(err))
	} else {
		log.Info("http shutdown complete")
	}
}
