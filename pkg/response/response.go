package response

import (
	"net/http"

	"github.com/labstack/echo/v5"

	"service-app/pkg/apperror"
)

// Response is the standard API response envelope.
type Response struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// Success sends a 200 OK response with data.
func Success(c *echo.Context, message string, data any) error {
	return c.JSON(http.StatusOK, Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// Created sends a 201 Created response with data.
func Created(c *echo.Context, message string, data any) error {
	return c.JSON(http.StatusCreated, Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// Error sends an error response. It extracts the HTTP status code and message
// from AppError if available, otherwise falls back to 500.
func Error(c *echo.Context, err error) error {
	code := apperror.HTTPCode(err)
	msg := apperror.HTTPMessage(err)

	return c.JSON(code, Response{
		Success: false,
		Message: msg,
	})
}

// ErrorWithCode sends an error response with a specific HTTP status code.
func ErrorWithCode(c *echo.Context, code int, message string) error {
	return c.JSON(code, Response{
		Success: false,
		Message: message,
	})
}
