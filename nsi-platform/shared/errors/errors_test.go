package errors

import (
	"net/http"
	"testing"
)

func TestErrNotFound(t *testing.T) {
	err := NewNotFound("user", "user-123")
	if err.Code != "NOT_FOUND" {
		t.Errorf("expected NOT_FOUND, got %s", err.Code)
	}
	if err.Message != "user not found: user-123" {
		t.Errorf("unexpected message: %s", err.Message)
	}
	if err.HTTPStatus != http.StatusNotFound {
		t.Errorf("expected 404, got %d", err.HTTPStatus)
	}
}

func TestErrValidation(t *testing.T) {
	err := NewValidation("email", "invalid format")
	if err.Code != "VALIDATION_ERROR" {
		t.Errorf("expected VALIDATION_ERROR, got %s", err.Code)
	}
	if err.Message != "email: invalid format" {
		t.Errorf("unexpected message: %s", err.Message)
	}
	if err.HTTPStatus != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", err.HTTPStatus)
	}
}

func TestErrUnauthorized(t *testing.T) {
	err := NewUnauthorized("missing x-user-id header")
	if err.Code != "UNAUTHORIZED" {
		t.Errorf("expected UNAUTHORIZED, got %s", err.Code)
	}
	if err.HTTPStatus != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", err.HTTPStatus)
	}
}

func TestErrInternal(t *testing.T) {
	err := NewInternal("database connection failed")
	if err.Code != "INTERNAL_ERROR" {
		t.Errorf("expected INTERNAL_ERROR, got %s", err.Code)
	}
	if err.HTTPStatus != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", err.HTTPStatus)
	}
}

func TestErrConflict(t *testing.T) {
	err := NewConflict("profile", "user-123")
	if err.Code != "CONFLICT" {
		t.Errorf("expected CONFLICT, got %s", err.Code)
	}
	if err.HTTPStatus != http.StatusConflict {
		t.Errorf("expected 409, got %d", err.HTTPStatus)
	}
}

func TestErrorImplementsErrorInterface(t *testing.T) {
	err := NewNotFound("test", "1")
	var e error = err
	if e.Error() == "" {
		t.Error("Error() should return non-empty string")
	}
}

func TestValidationErrors(t *testing.T) {
	errs := NewValidationErrors()
	errs.Add("email", "is required")
	errs.Add("age", "must be positive")

	if errs.HasErrors() != true {
		t.Error("expected HasErrors() to be true")
	}
	if len(errs.Errors) != 2 {
		t.Errorf("expected 2 errors, got %d", len(errs.Errors))
	}
	if errs.Error() == "" {
		t.Error("Error() should return non-empty string")
	}
}

func TestValidationErrorsEmpty(t *testing.T) {
	errs := NewValidationErrors()
	if errs.HasErrors() != false {
		t.Error("expected HasErrors() to be false")
	}
}
