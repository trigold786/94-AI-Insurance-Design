package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/trigold786/94-AI-Insurance-Design/shared/middleware"
	"github.com/trigold786/94-AI-Insurance-Design/shared/models"
)

type mockRightsRepo struct {
	records []models.PaymentRecord
	alerts  []models.Alert
	err     error
}

func (m *mockRightsRepo) GetPaymentRecords(ctx context.Context, userID string) ([]models.PaymentRecord, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.records, nil
}

func (m *mockRightsRepo) GetAlerts(ctx context.Context, userID string, unreadOnly bool) ([]models.Alert, error) {
	if m.err != nil {
		return nil, m.err
	}
	if unreadOnly {
		var unread []models.Alert
		for _, a := range m.alerts {
			if !a.IsRead {
				unread = append(unread, a)
			}
		}
		return unread, nil
	}
	return m.alerts, nil
}

func (m *mockRightsRepo) MarkAlertRead(ctx context.Context, alertID, userID string) error {
	return m.err
}
func (m *mockRightsRepo) CreateAlert(ctx context.Context, alert *models.Alert) error {
	m.alerts = append(m.alerts, *alert)
	return nil
}
func (m *mockRightsRepo) GetDueSoonPayments(ctx context.Context, withinDays int) ([]models.PaymentRecord, error) {
	return nil, nil
}
func (m *mockRightsRepo) UpsertPaymentRecord(ctx context.Context, record *models.PaymentRecord) error {
	if m.err != nil {
		return m.err
	}
	m.records = append(m.records, *record)
	return nil
}

func TestPaymentStatusHandler(t *testing.T) {
	t.Run("returns payment summary", func(t *testing.T) {
		repo := &mockRightsRepo{
			records: []models.PaymentRecord{
				{RecordID: "r1", PolicyType: "pension", Month: "2026-01", Amount: 1000, Status: "paid", DueDate: "2026-01-15"},
				{RecordID: "r2", PolicyType: "medical", Month: "2026-01", Amount: 500, Status: "pending", DueDate: "2026-01-15"},
				{RecordID: "r3", PolicyType: "pension", Month: "2026-02", Amount: 1000, Status: "missed", DueDate: "2026-02-15"},
			},
		}

		handler := middleware.AuthMiddleware(testJWTSecret)(PaymentStatusHandler(repo))
		req := httptest.NewRequest("GET", "/v1/rights/payment-status", nil)
		setAuth(req, "user-1")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}

		var resp struct {
			Code int                    `json:"code"`
			Data map[string]interface{} `json:"data"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to decode: %v", err)
		}
		if resp.Data["paid_count"].(float64) != 1 {
			t.Errorf("expected 1 paid, got %v", resp.Data["paid_count"])
		}
		if resp.Data["missed_count"].(float64) != 1 {
			t.Errorf("expected 1 missed, got %v", resp.Data["missed_count"])
		}
	})

	t.Run("returns 401 without user", func(t *testing.T) {
		handler := PaymentStatusHandler(&mockRightsRepo{})
		req := httptest.NewRequest("GET", "/v1/rights/payment-status", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d", w.Code)
		}
	})

	t.Run("returns empty summary when no records", func(t *testing.T) {
		handler := middleware.AuthMiddleware(testJWTSecret)(PaymentStatusHandler(&mockRightsRepo{records: []models.PaymentRecord{}}))
		req := httptest.NewRequest("GET", "/v1/rights/payment-status", nil)
		setAuth(req, "user-1")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		var resp struct {
			Data map[string]interface{} `json:"data"`
		}
		json.Unmarshal(w.Body.Bytes(), &resp)
		if resp.Data["total_records"].(float64) != 0 {
			t.Errorf("expected 0 records, got %v", resp.Data["total_records"])
		}
	})
}

func TestAlertListHandler(t *testing.T) {
	t.Run("returns all alerts", func(t *testing.T) {
		repo := &mockRightsRepo{
			alerts: []models.Alert{
				{AlertID: "a1", AlertType: "disconnection_risk", Severity: "high", Title: "断缴风险", IsRead: false},
				{AlertID: "a2", AlertType: "policy_change", Severity: "low", Title: "政策更新", IsRead: true},
			},
		}

		handler := middleware.AuthMiddleware(testJWTSecret)(AlertListHandler(repo))
		req := httptest.NewRequest("GET", "/v1/rights/alerts", nil)
		setAuth(req, "user-1")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}

		var resp struct {
			Code int            `json:"code"`
			Data []models.Alert `json:"data"`
		}
		json.Unmarshal(w.Body.Bytes(), &resp)
		if len(resp.Data) != 2 {
			t.Errorf("expected 2 alerts, got %d", len(resp.Data))
		}
	})

	t.Run("filters unread only", func(t *testing.T) {
		repo := &mockRightsRepo{
			alerts: []models.Alert{
				{AlertID: "a1", AlertType: "disconnection_risk", IsRead: false},
				{AlertID: "a2", AlertType: "policy_change", IsRead: true},
			},
		}

		handler := middleware.AuthMiddleware(testJWTSecret)(AlertListHandler(repo))
		req := httptest.NewRequest("GET", "/v1/rights/alerts?unread_only=true", nil)
		setAuth(req, "user-1")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		var resp struct {
			Data []models.Alert `json:"data"`
		}
		json.Unmarshal(w.Body.Bytes(), &resp)
		if len(resp.Data) != 1 {
			t.Errorf("expected 1 unread alert, got %d", len(resp.Data))
		}
	})

	t.Run("mark alert as read", func(t *testing.T) {
		handler := middleware.AuthMiddleware(testJWTSecret)(MarkAlertReadHandler(&mockRightsRepo{}))
		req := httptest.NewRequest("PUT", "/v1/rights/alerts/read?alert_id=a1", nil)
		setAuth(req, "user-1")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})

	t.Run("mark alert requires alert_id", func(t *testing.T) {
		handler := middleware.AuthMiddleware(testJWTSecret)(MarkAlertReadHandler(&mockRightsRepo{}))
		req := httptest.NewRequest("PUT", "/v1/rights/alerts/read", nil)
		setAuth(req, "user-1")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})
}

func TestNewAlertScheduler(t *testing.T) {
	scheduler := NewAlertScheduler(&mockRightsRepo{})
	if scheduler == nil {
		t.Fatal("expected non-nil scheduler")
	}

	scheduler.Start(time.Hour)
	scheduler.Stop()
}
