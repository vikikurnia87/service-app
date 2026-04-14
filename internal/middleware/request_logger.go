package middleware

import (
	"log/slog"
	"time"

	"github.com/labstack/echo/v5"
)

// RequestLogger returns an Echo middleware that logs every request using slog.
func RequestLogger(logger *slog.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			start := time.Now()

			err := next(c)

			req := c.Request()
			latency := time.Since(start)

			attrs := []any{
				"method", req.Method,
				"path", req.URL.Path,
				"status", c.Response().(*echo.Response).Status,
				"latency", latency.String(),
				"remote_ip", req.RemoteAddr,
			}

			if err != nil {
				attrs = append(attrs, "error", err.Error())
				logger.Error("request failed", attrs...)
			} else {
				logger.Info("request completed", attrs...)
			}

			return err
		}
	}
}
