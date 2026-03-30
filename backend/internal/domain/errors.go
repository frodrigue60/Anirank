package domain

import "errors"

var (
	ErrNotFound          = errors.New("resource not found")
	ErrInternalServer    = errors.New("internal server error")
	ErrUnauthorized      = errors.New("unauthorized request")
	ErrForbidden         = errors.New("forbidden resource access")
	ErrInvalidValidation = errors.New("validation failed")
	ErrConflict          = errors.New("resource conflict")
)

// AppError is a custom error struct that holds HTTP status codes for easier error handling in HTTP responses.
type AppError struct {
	Code    int
	Message string
	Err     error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

func NewAppError(code int, message string, err error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// Response is a standardized JSON response structure for the API.
type Response struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
}
