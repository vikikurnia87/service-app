package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hibiken/asynq"
	"github.com/labstack/echo/v5"

	"service-app/config"
	appCache "service-app/internal/cache"
	"service-app/internal/handler"
	"service-app/internal/repository"
	"service-app/internal/service"
	"service-app/internal/worker"
	"service-app/pkg/database"
	appRedis "service-app/pkg/redis"
	"service-app/routes"
)

func main() {
	// -------------------------------------------------------------------
	// 1. Logger
	// -------------------------------------------------------------------
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	ctx := context.Background()

	// -------------------------------------------------------------------
	// 2. Config
	// -------------------------------------------------------------------
	cfg, err := config.Load()
	if err != nil {
		logger.Error("failed to load config", "error", err)
		os.Exit(1)
	}
	logger.Info("config loaded successfully",
		"app_name", cfg.AppName,
		"env", cfg.Server.Env,
	)

	// -------------------------------------------------------------------
	// 3. Database (PostgreSQL + Bun) — REQUIRED
	// -------------------------------------------------------------------
	db, err := database.NewPostgresDB(cfg.DB, cfg.Server.Env, logger)
	if err != nil {
		logger.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer database.Close(db)

	// -------------------------------------------------------------------
	// 4. Redis — OPTIONAL (app runs without it)
	// -------------------------------------------------------------------
	appRedis.InitRedisDatabase(ctx, cfg.Redis, logger)
	defer appRedis.RedisClose(ctx, logger)

	// Use real Redis cache or no-op fallback
	var cache appCache.Cache
	if appRedis.IsReady() {
		cache = appCache.NewRedisCache(appRedis.GetClient())
	} else {
		cache = appCache.NewNoopCache()
		logger.Info("using no-op cache (Redis not available)")
	}

	// -------------------------------------------------------------------
	// 5. Asynq (background jobs) — OPTIONAL (requires Redis)
	// -------------------------------------------------------------------
	var asynqServer *asynq.Server
	if appRedis.IsReady() {
		asynqClient := worker.NewAsynqClient(cfg.Redis)
		defer asynqClient.Close()

		asynqServer = worker.NewAsynqServer(cfg.Redis, cfg.Asynq, logger)
		asynqMux := asynq.NewServeMux()
		worker.RegisterHandlers(asynqMux, logger)

		go func() {
			logger.Info("starting asynq worker", "concurrency", cfg.Asynq.Concurrency)
			if err := asynqServer.Run(asynqMux); err != nil {
				logger.Error("asynq worker stopped", "error", err)
			}
		}()
	} else {
		logger.Info("asynq worker disabled (Redis not available)")
	}

	// -------------------------------------------------------------------
	// 6. Dependency Injection: Repository → Service → Handler
	// -------------------------------------------------------------------
	userRepo := repository.NewUserRepository(db)
	userSvc := service.NewUserService(userRepo, cache, logger)
	userHandler := handler.NewUserHandler(userSvc, logger)

	roleRepo := repository.NewRoleRepository(db)
	roleSvc := service.NewRoleService(roleRepo, cache, logger)
	roleHandler := handler.NewRoleHandler(roleSvc, logger)

	healthHandler := handler.NewHealthHandler()

	// -------------------------------------------------------------------
	// 7. Echo Server
	// -------------------------------------------------------------------
	e := echo.New()
	e.Logger = logger

	routes.RegisterRoutes(e, routes.Handlers{
		Health: healthHandler,
		User:   userHandler,
		Role:   roleHandler,
	}, logger)

	// -------------------------------------------------------------------
	// 8. Graceful Shutdown
	// -------------------------------------------------------------------
	shutdownCtx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	sc := echo.StartConfig{
		Address:         cfg.Server.Address(),
		GracefulTimeout: 10 * time.Second,
	}

	logger.Info("starting server", "address", cfg.Server.Address())
	if err := sc.Start(shutdownCtx, e); err != nil {
		logger.Error("server stopped", "error", err)
	}

	// Shutdown asynq if it was started
	if asynqServer != nil {
		asynqServer.Shutdown()
	}
	logger.Info("application shutdown complete")
}
