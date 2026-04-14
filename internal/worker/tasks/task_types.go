package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/hibiken/asynq"
)

// Task type constants. Add new task types here.
const (
	TypeWelcomeEmail = "email:welcome"
)

// WelcomeEmailPayload holds the data needed to send a welcome email.
type WelcomeEmailPayload struct {
	UserID int64  `json:"user_id"`
	Email  string `json:"email"`
}

// NewWelcomeEmailTask creates a new welcome email task ready for enqueuing.
func NewWelcomeEmailTask(userID int64, email string) (*asynq.Task, error) {
	payload, err := json.Marshal(WelcomeEmailPayload{
		UserID: userID,
		Email:  email,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal welcome email payload: %w", err)
	}
	return asynq.NewTask(TypeWelcomeEmail, payload), nil
}

// HandleWelcomeEmailTask returns a handler function for the welcome email task.
func HandleWelcomeEmailTask(logger *slog.Logger) func(ctx context.Context, t *asynq.Task) error {
	return func(ctx context.Context, t *asynq.Task) error {
		var p WelcomeEmailPayload
		if err := json.Unmarshal(t.Payload(), &p); err != nil {
			return fmt.Errorf("failed to unmarshal welcome email payload: %w", err)
		}

		logger.Info("processing welcome email task",
			"user_id", p.UserID,
			"email", p.Email,
		)

		// TODO: Implement actual email sending logic here.

		logger.Info("welcome email task completed",
			"user_id", p.UserID,
			"email", p.Email,
		)
		return nil
	}
}
