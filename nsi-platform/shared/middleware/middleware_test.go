package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAuthMiddlewareSetsUserID(t *testing.T) {
	handler := AuthMiddleware("")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value(ContextKeyUserID)
		if userID == nil {
			t.Error("expected user_id in context")
		}
		if userID.(string) != "user-abc-123" {
			t.Errorf("expected user-abc-123, got %v", userID)
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/v1/profile", nil)
	req.Header.Set("x-user-id", "user-abc-123")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestAuthMiddlewareMissingHeader(t *testing.T) {
	handler := AuthMiddleware("")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("should not reach handler when header is missing")
	}))

	req := httptest.NewRequest("GET", "/v1/profile", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestAuthMiddlewareEmptyHeader(t *testing.T) {
	handler := AuthMiddleware("")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("should not reach handler when header is empty")
	}))

	req := httptest.NewRequest("GET", "/v1/profile", nil)
	req.Header.Set("x-user-id", "")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestRecoveryMiddlewareCatchesPanic(t *testing.T) {
	var gotCode int
	handler := RecoveryMiddleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	}))

	req := httptest.NewRequest("GET", "/panic", nil)
	w := httptest.NewRecorder()
	gotCode = http.StatusOK

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
	_ = gotCode
}

func TestRecoveryMiddlewareNormalRequest(t *testing.T) {
	handler := RecoveryMiddleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if w.Body.String() != "ok" {
		t.Errorf("expected 'ok', got '%s'", w.Body.String())
	}
}

func TestChainMiddleware(t *testing.T) {
	handler := Chain(
		RecoveryMiddleware(),
		AuthMiddleware(""),
	)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value(ContextKeyUserID)
		if userID == nil || userID.(string) != "test-user" {
			t.Errorf("expected test-user in context, got %v", userID)
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/v1/profile", nil)
	req.Header.Set("x-user-id", "test-user")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}
