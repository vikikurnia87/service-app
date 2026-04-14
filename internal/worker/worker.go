package worker

import (
	"log/slog"

	"github.com/hibiken/asynq"

	"service-app/config"
	"service-app/internal/worker/tasks"
)

// NewAsynqServer creates a new Asynq worker server.
func NewAsynqServer(cfg config.RedisConfig, asynqCfg config.AsynqConfig, logger *slog.Logger) *asynq.Server {
	return asynq.NewServer(
		asynq.RedisClientOpt{
			Addr:     cfg.Addr(),
			Password: cfg.Password,
			DB:       cfg.DB,
		},
		asynq.Config{
			Concurrency: asynqCfg.Concurrency,
			// Logger:      slog.NewLogLogger(logger.Handler(), slog.LevelInfo),
		},
	)
}

// NewAsynqClient creates a new Asynq client for enqueuing tasks.
func NewAsynqClient(cfg config.RedisConfig) *asynq.Client {
	return asynq.NewClient(asynq.RedisClientOpt{
		Addr:     cfg.Addr(),
		Password: cfg.Password,
		DB:       cfg.DB,
	})
}

// RegisterHandlers registers all task handlers to the Asynq mux.
func RegisterHandlers(mux *asynq.ServeMux, logger *slog.Logger) {
	mux.HandleFunc(tasks.TypeWelcomeEmail, tasks.HandleWelcomeEmailTask(logger))
}
