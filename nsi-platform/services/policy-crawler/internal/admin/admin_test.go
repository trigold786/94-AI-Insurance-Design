package admin

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestListClaimsHandler(t *testing.T) {
	handler := ListClaimsHandler(nil)

	req := httptest.NewRequest("GET", "/admin/claims?status=pending_review", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), `"claims"`) {
		t.Errorf("expected claims field in response, got %s", w.Body.String())
	}
}

func TestListClaimsHandlerInvalidStatus(t *testing.T) {
	handler := ListClaimsHandler(nil)

	req := httptest.NewRequest("GET", "/admin/claims?status=invalid", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid status, got %d", w.Code)
	}
}

func TestListClaimsHandlerMissingStatus(t *testing.T) {
	handler := ListClaimsHandler(nil)

	req := httptest.NewRequest("GET", "/admin/claims", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestUpdateClaimHandlerValid(t *testing.T) {
	handler := UpdateClaimHandler(nil)

	body := `{"status":"verified","confidence_score":0.95}`
	req := httptest.NewRequest("PUT", "/admin/claims/CLM-001", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestUpdateClaimHandlerInvalidStatus(t *testing.T) {
	handler := UpdateClaimHandler(nil)

	body := `{"status":"invalid_status"}`
	req := httptest.NewRequest("PUT", "/admin/claims/CLM-001", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestUpdateClaimHandlerInvalidJSON(t *testing.T) {
	handler := UpdateClaimHandler(nil)

	req := httptest.NewRequest("PUT", "/admin/claims/CLM-001", strings.NewReader(`not json`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestUpdateClaimHandlerNoID(t *testing.T) {
	handler := UpdateClaimHandler(nil)

	body := `{"status":"verified"}`
	req := httptest.NewRequest("PUT", "/admin/claims/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}
