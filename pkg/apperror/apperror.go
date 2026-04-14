package apperror

import (
	"errors"
	"fmt"
	"net/http"
)

// AppError represents an application-level error with HTTP status code.
type AppError struct {
	Code    int    // HTTP status code
	Message string // User-facing message
	Err     error  // Wrapped internal error
}

// Error implements the error interface.
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// Unwrap returns the wrapped error for errors.Is/As compatibility.
func (e *AppError) Unwrap() error {
	return e.Err
}

// New creates a new AppError with the given code and message.
func New(code int, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
	}
}

// Wrap creates a new AppError wrapping an existing error.
func Wrap(code int, message string, err error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// Sentinel errors for common cases.
var (
	ErrNotFound     = New(http.StatusNotFound, "resource not found")
	ErrBadRequest   = New(http.StatusBadRequest, "bad request")
	ErrUnauthorized = New(http.StatusUnauthorized, "unauthorized")
	ErrForbidden    = New(http.StatusForbidden, "forbidden")
	ErrInternal     = New(http.StatusInternalServerError, "internal server error")
)

// NewNotFound creates a not-found error with a custom message.
func NewNotFound(message string) *AppError {
	return New(http.StatusNotFound, message)
}

// NewBadRequest creates a bad-request error with a custom message.
func NewBadRequest(message string) *AppError {
	return New(http.StatusBadRequest, message)
}

// NewInternal creates an internal error wrapping a cause.
func NewInternal(err error) *AppError {
	return Wrap(http.StatusInternalServerError, "internal server error", err)
}

// HTTPCode extracts the HTTP status code from an error if it is an AppError.
// Returns 500 for non-AppError types.
func HTTPCode(err error) int {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Code
	}
	return http.StatusInternalServerError
}

// HTTPMessage extracts the user-facing message from an error if it is an AppError.
// Returns a generic message for non-AppError types.
func HTTPMessage(err error) string {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Message
	}
	return "internal server error"
}
