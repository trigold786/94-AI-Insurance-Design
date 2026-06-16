package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/trigold786/94-AI-Insurance-Design/shared/errors"
	"github.com/trigold786/94-AI-Insurance-Design/shared/middleware"
	"github.com/trigold786/94-AI-Insurance-Design/shared/models"
)

type mockProfileRepo struct {
	profile *models.UserProfile
	err     error
}

func (m *mockProfileRepo) GetByUserID(ctx context.Context, userID string) (*models.UserProfile, error) {
	return m.profile, m.err
}

func (m *mockProfileRepo) Upsert(ctx context.Context, profile *models.UserProfile) error {
	return m.err
}

func TestHealthCheckHandler(t *testing.T) {
	handler := HealthCheckHandler()

	req := httptest.NewRequest("GET", "/healthz", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/json; charset=utf-8" {
		t.Errorf("expected JSON content type, got %s", ct)
	}
	if !strings.Contains(w.Body.String(), `"status":"ok"`) {
		t.Errorf("expected status ok in response, got %s", w.Body.String())
	}
}

func TestGetProfileHandlerSuccess(t *testing.T) {
	profile := &models.UserProfile{
		UserID:              "user-123",
		Age:                 30,
		Gender:              "male",
		EmploymentStatus:    "flexible",
		SocialSecurityYears: 8,
	}
	repo := &mockProfileRepo{profile: profile}
	handler := middleware.AuthMiddleware(testJWTSecret)(GetProfileHandler(repo))

	req := httptest.NewRequest("GET", "/v1/profile", nil)
	setAuth(req, "user-123")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), `"user_id":"user-123"`) {
		t.Errorf("expected user-123 in response, got %s", w.Body.String())
	}
}

func TestGetProfileHandlerNotFound(t *testing.T) {
	repo := &mockProfileRepo{err: errors.NewNotFound("user profile", "user-999")}
	handler := middleware.AuthMiddleware(testJWTSecret)(GetProfileHandler(repo))

	req := httptest.NewRequest("GET", "/v1/profile", nil)
	setAuth(req, "user-999")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestGetProfileHandlerUnauthenticated(t *testing.T) {
	repo := &mockProfileRepo{}
	handler := middleware.AuthMiddleware(testJWTSecret)(GetProfileHandler(repo))

	req := httptest.NewRequest("GET", "/v1/profile", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestUpdateProfileHandlerSuccess(t *testing.T) {
	repo := &mockProfileRepo{}
	handler := middleware.AuthMiddleware(testJWTSecret)(UpdateProfileHandler(repo))

	body := `{"age":30,"gender":"male","employment_status":"flexible"}`
	req := httptest.NewRequest("PUT", "/v1/profile", strings.NewReader(body))
	setAuth(req, "user-123")
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestUpdateProfileHandlerInvalidJSON(t *testing.T) {
	repo := &mockProfileRepo{}
	handler := middleware.AuthMiddleware(testJWTSecret)(UpdateProfileHandler(repo))

	req := httptest.NewRequest("PUT", "/v1/profile", strings.NewReader(`invalid json`))
	setAuth(req, "user-123")
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}
