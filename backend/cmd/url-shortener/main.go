package main

import (
	"context"
	"errors"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/wb-go/wbf/dbpg"
	"github.com/wb-go/wbf/redis"
	"github.com/wb-go/wbf/zlog"

	"github.com/aliskhannn/url-shortener/internal/api/handlers/analytics"
	"github.com/aliskhannn/url-shortener/internal/api/handlers/link"
	"github.com/aliskhannn/url-shortener/internal/api/router"
	"github.com/aliskhannn/url-shortener/internal/api/server"
	"github.com/aliskhannn/url-shortener/internal/config"
	analyticsrepo "github.com/aliskhannn/url-shortener/internal/repository/analytics"
	linkrepo "github.com/aliskhannn/url-shortener/internal/repository/link"
	analyticssvc "github.com/aliskhannn/url-shortener/internal/service/analytics"
	linksvc "github.com/aliskhannn/url-shortener/internal/service/link"
)

func main() {
	// Setup context to handle SIGINT and SIGTERM for graceful shutdown.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Initialize logger and configuration.
	zlog.Init()
	cfg := config.Must()
	val := validator.New()

	// Connect to PostgreSQL master and slave databases.
	opts := &dbpg.Options{
		MaxOpenConns:    cfg.Database.MaxOpenConns,
		MaxIdleConns:    cfg.Database.MaxIdleConns,
		ConnMaxLifetime: cfg.Database.ConnMaxLifetime,
	}

	slaveDNSs := make([]string, 0, len(cfg.Database.Slaves))

	for _, s := range cfg.Database.Slaves {
		slaveDNSs = append(slaveDNSs, s.DSN())
	}
	zlog.Logger.Info().Msgf("db url: %s", cfg.Database.Master.DSN())
	db, err := dbpg.New(cfg.Database.Master.DSN(), slaveDNSs, opts)
	if err != nil {
		zlog.Logger.Fatal().Err(err).Msg("failed to connect to database")
	}

	// Connect to Redis
	dbNum, err := strconv.Atoi(cfg.Redis.Database)
	if err != nil {
		zlog.Logger.Fatal().Err(err).Msg("failed to parse redis database")
	}

	zlog.Logger.Info().Msgf("redis config: %s, %s, %d", cfg.Redis.Address, cfg.Redis.Password, dbNum)
	rdb := redis.New(cfg.Redis.Address, cfg.Redis.Password, dbNum)

	if err = rdb.Ping(ctx).Err(); err != nil {
		zlog.Logger.Fatal().Err(err).Msg("failed to connect to redis")
	}

	// Initialize link and analytics repository, service and handlers.
	linkRepo := linkrepo.NewRepository(db)
	analyticsRepo := analyticsrepo.NewRepository(db)

	linkService := linksvc.NewService(linkRepo, rdb)
	analyticsService := analyticssvc.NewService(analyticsRepo, rdb)

	linkHandler := link.NewHandler(linkService, analyticsService, val, cfg)
	analyticsHandler := analytics.NewHandler(analyticsService, cfg)

	r := router.New(linkHandler, analyticsHandler)
	s := server.New(cfg.Server.HTTPPort, r)
	go func() {
		if err := s.ListenAndServe(); err != nil {
			zlog.Logger.Fatal().Err(err).Msg("failed to start server")
		}
	}()

	// Wait for shutdown signal.
	<-ctx.Done()
	zlog.Logger.Info().Msg("shutdown signal received")

	// Graceful shutdown with timeout.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	zlog.Logger.Info().Msg("shutting down server")
	if err := s.Shutdown(shutdownCtx); err != nil {
		zlog.Logger.Error().Err(err).Msg("failed to shutdown server")
	}
	if errors.Is(shutdownCtx.Err(), context.DeadlineExceeded) {
		zlog.Logger.Info().Msg("timeout exceeded, forcing shutdown")
	}

	// Close master and slave databases.
	if err := db.Master.Close(); err != nil {
		zlog.Logger.Printf("failed to close master DB: %v", err)
	}
	for i, s := range db.Slaves {
		if err := s.Close(); err != nil {
			zlog.Logger.Printf("failed to close slave DB %d: %v", i, err)
		}
	}
}
