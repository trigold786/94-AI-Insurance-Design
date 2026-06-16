package admin

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/trigold786/94-AI-Insurance-Design/shared/models"
)

func TestIngestPolicyHandler(t *testing.T) {
	t.Run("imports structured policy text", func(t *testing.T) {
		store := &mockStore{}
		handler := IngestPolicyHandler(store)

		text := strings.Join([]string{
			"政策ID: TEST-001",
			"地区代码: 310000",
			"政策类型: subsidy",
			"适用人群: flexible_employment, 4050",
			"计算方法: 基数*50%",
			"补贴下限: 300",
			"补贴上限: 800",
			"补贴时长: 24",
			"生效日期: 2025-01-01",
			"失效日期: 2026-12-31",
			"认定条件:",
			"  - name: 灵活就业登记",
			"    desc: 已办理灵活就业登记",
			"    tag: flexible_employment",
			"必需材料:",
			"  - name: 身份证",
			"    desc: 原件及复印件",
			"    source: user",
		}, "\n")

		body, _ := json.Marshal(map[string]string{"text": text})
		req := httptest.NewRequest("POST", "/admin/ingest", strings.NewReader(string(body)))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}

		var resp struct {
			Code    int                    `json:"code"`
			Message string                 `json:"message"`
			Data    map[string]interface{} `json:"data"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to decode: %v", err)
		}
		if resp.Code != 0 {
			t.Fatalf("expected code 0, got %d: %s", resp.Code, resp.Message)
		}
		if resp.Data["policy_id"] != "TEST-001" {
			t.Errorf("expected policy_id TEST-001, got %v", resp.Data["policy_id"])
		}
		if resp.Data["status"] != "pending_review" {
			t.Errorf("expected status pending_review, got %v", resp.Data["status"])
		}
		if store.ingested == nil {
			t.Fatal("expected claim to be ingested")
		}
	})

	t.Run("rejects empty text", func(t *testing.T) {
		handler := IngestPolicyHandler(&mockStore{})
		body, _ := json.Marshal(map[string]string{"text": ""})
		req := httptest.NewRequest("POST", "/admin/ingest", strings.NewReader(string(body)))
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})

	t.Run("rejects missing fields", func(t *testing.T) {
		handler := IngestPolicyHandler(&mockStore{})
		body, _ := json.Marshal(map[string]string{"text": "invalid text without required fields"})
		req := httptest.NewRequest("POST", "/admin/ingest", strings.NewReader(string(body)))
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})
}

type mockStore struct {
	ingested *bool
}

func (m *mockStore) ListByStatus(status string, regionCode string, sourceID string, policyType string, sourceLevel string) ([]models.PolicyClaim, error) { return nil, nil }
func (m *mockStore) UpdateStatus(claimID, status string, confidence float64) error { return nil }
func (m *mockStore) Ingest(claim *models.PolicyClaim) error { m.ingested = &[]bool{true}[0]; return nil }
func (m *mockStore) GetClaimByID(claimID string) (*models.PolicyClaim, error) { return nil, nil }
func (m *mockStore) SearchSimilarClaims(claimID string, limit int) ([]models.PolicyClaim, error) { return nil, nil }
func (m *mockStore) UpdateClaimFields(claimID string, fields map[string]interface{}) error { return nil }