package errors

import (
	"fmt"
	"net/http"
)

type AppError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	StatusCode int    `json:"-"`
}

func (e *AppError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func NewAppError(code string, message string, statusCode int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
	}
}

var (
	ErrNotFound       = NewAppError("ERR-404-NOT-FOUND", "Resource not found", http.StatusNotFound)
	ErrInternalServer = NewAppError("ERR-500-INTERNAL", "Internal server error", http.StatusInternalServerError)
	ErrBadRequest     = NewAppError("ERR-400-BAD-REQUEST", "Invalid request parameters", http.StatusBadRequest)
	ErrFoodNotFound   = NewAppError("ERR-404-FOOD-NOT-FOUND", "Food item out of stock or not found", http.StatusNotFound)
	ErrUnauthorized   = NewAppError("ERR-401-UNAUTHORIZED", "Unauthorized access", http.StatusUnauthorized)
)
