package errors

import (
	"fmt"
	"net/http"
)

type AppError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	HTTPStatus int    `json:"-"`
}

func (e *AppError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func NewNotFound(resource, id string) *AppError {
	return &AppError{
		Code:       "NOT_FOUND",
		Message:    fmt.Sprintf("%s not found: %s", resource, id),
		HTTPStatus: http.StatusNotFound,
	}
}

func NewValidation(field, reason string) *AppError {
	return &AppError{
		Code:       "VALIDATION_ERROR",
		Message:    fmt.Sprintf("%s: %s", field, reason),
		HTTPStatus: http.StatusBadRequest,
	}
}

func NewUnauthorized(reason string) *AppError {
	return &AppError{
		Code:       "UNAUTHORIZED",
		Message:    reason,
		HTTPStatus: http.StatusUnauthorized,
	}
}

func NewInternal(detail string) *AppError {
	return &AppError{
		Code:       "INTERNAL_ERROR",
		Message:    detail,
		HTTPStatus: http.StatusInternalServerError,
	}
}

func NewConflict(resource, id string) *AppError {
	return &AppError{
		Code:       "CONFLICT",
		Message:    fmt.Sprintf("%s already exists: %s", resource, id),
		HTTPStatus: http.StatusConflict,
	}
}

type ValidationErrors struct {
	Errors map[string]string `json:"errors"`
}

func NewValidationErrors() *ValidationErrors {
	return &ValidationErrors{Errors: make(map[string]string)}
}

func (ve *ValidationErrors) Add(field, message string) {
	ve.Errors[field] = message
}

func (ve *ValidationErrors) HasErrors() bool {
	return len(ve.Errors) > 0
}

func (ve *ValidationErrors) Error() string {
	return fmt.Sprintf("validation errors: %v", ve.Errors)
}
