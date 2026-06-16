package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/trigold786/94-AI-Insurance-Design/shared/middleware"
	"github.com/trigold786/94-AI-Insurance-Design/shared/models"
)

type mockPlanRepo struct {
	savedPlan *models.PlanSnapshot
	err       error
}

func (m *mockPlanRepo) Save(ctx context.Context, plan *models.PlanSnapshot) error {
	if m.err != nil {
		return m.err
	}
	m.savedPlan = plan
	return nil
}

func (m *mockPlanRepo) GetByID(ctx context.Context, planID string) (*models.PlanSnapshot, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.savedPlan != nil {
		return m.savedPlan, nil
	}
	return nil, nil
}

const mockLLMResponseBody = `===FREE_FORM_START===
建议您选择中等缴费基数方案...
===FREE_FORM_END===
===STRUCTURED_START===
{
  "summary": "推荐中等基数方案",
  "schemes": [
    {"name":"方案A","description":"中等基数","monthly_cost":800,"annual_subsidy":1200,"projected_pension":2500,"total_cost":9600,"contribution_base":6000,"analysis":"适合您的收入水平","applicable_policies":["CLM-001"]}
  ],
  "policy_references": [
    {"claim_id":"CLM-001","policy_title":"灵活就业人员社保补贴","document_number":"沪人社发[2024]1号","policy_url":"https://example.com","relevant_excerpt":"月缴费基数下限为社平工资60%","how_applied":"用于确定缴费基数下限"}
  ],
  "recommendation": {"recommended_scheme":"方案A","reasoning":"性价比最优"}
}
===STRUCTURED_END===`

func newMockLLMGateway(statusCode int, responseBody string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chat" {
			w.WriteHeader(404)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(LLMChatResponse{
			Code:    statusCode,
			Content: responseBody,
		})
	}))
}

func TestGeneratePlanHandlerLLMSuccess(t *testing.T) {
	gw := newMockLLMGateway(200, mockLLMResponseBody)
	defer gw.Close()

	repo := &mockPlanRepo{}
	handler := GeneratePlanHandler(gw.URL, "", repo, nil, nil)

	body := `{"age":30,"gender":"male","employment":"flexible","monthly_budget":3000}`
	req := httptest.NewRequest("POST", "/v1/plans/generate", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), middleware.ContextKeyUserID, "user-1")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	if repo.savedPlan == nil {
		t.Fatal("expected plan to be saved")
	}
	if repo.savedPlan.FreeFormText == "" {
		t.Error("expected free form text in saved plan")
	}
	if len(repo.savedPlan.StructuredSchemes) != 1 {
		t.Errorf("expected 1 structured scheme, got %d", len(repo.savedPlan.StructuredSchemes))
	}
	if repo.savedPlan.StructuredSchemes[0].Name != "方案A" {
		t.Errorf("expected scheme name '方案A', got %s", repo.savedPlan.StructuredSchemes[0].Name)
	}
	if repo.savedPlan.Recommendation != "方案A" {
		t.Errorf("expected recommendation '方案A', got %s", repo.savedPlan.Recommendation)
	}
	if len(repo.savedPlan.PolicyReferences) != 1 {
		t.Errorf("expected 1 policy reference, got %d", len(repo.savedPlan.PolicyReferences))
	}
	if repo.savedPlan.TotalCost != 9600 {
		t.Errorf("expected total cost 9600, got %f", repo.savedPlan.TotalCost)
	}
	if repo.savedPlan.TotalSubsidy != 1200 {
		t.Errorf("expected total subsidy 1200, got %f", repo.savedPlan.TotalSubsidy)
	}
}

func TestGeneratePlanHandlerInvalidAge(t *testing.T) {
	handler := GeneratePlanHandler("http://unused", "", nil, nil, nil)

	body := `{"age":-5,"gender":"male","monthly_budget":1000}`
	req := httptest.NewRequest("POST", "/v1/plans/generate", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestGeneratePlanHandlerInvalidGender(t *testing.T) {
	handler := GeneratePlanHandler("http://unused", "", nil, nil, nil)

	body := `{"age":30,"gender":"other","monthly_budget":1000}`
	req := httptest.NewRequest("POST", "/v1/plans/generate", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestGeneratePlanHandlerInvalidJSON(t *testing.T) {
	handler := GeneratePlanHandler("http://unused", "", nil, nil, nil)

	req := httptest.NewRequest("POST", "/v1/plans/generate", strings.NewReader(`not json`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestGeneratePlanHandlerLLMError(t *testing.T) {
	gw := newMockLLMGateway(500, "internal error")
	defer gw.Close()

	repo := &mockPlanRepo{}
	handler := GeneratePlanHandler(gw.URL, "", repo, nil, nil)

	body := `{"age":30,"gender":"male","employment":"flexible","monthly_budget":3000}`
	req := httptest.NewRequest("POST", "/v1/plans/generate", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), middleware.ContextKeyUserID, "user-1")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d: %s", w.Code, w.Body.String())
	}
	if repo.savedPlan != nil {
		t.Error("expected no plan to be saved on LLM error")
	}
}

func TestGeneratePlanHandlerParseError(t *testing.T) {
	gw := newMockLLMGateway(200, "this is just plain text without any markers")
	defer gw.Close()

	repo := &mockPlanRepo{}
	handler := GeneratePlanHandler(gw.URL, "", repo, nil, nil)

	body := `{"age":30,"gender":"male","employment":"flexible","monthly_budget":3000}`
	req := httptest.NewRequest("POST", "/v1/plans/generate", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), middleware.ContextKeyUserID, "user-1")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d: %s", w.Code, w.Body.String())
	}
	if repo.savedPlan != nil {
		t.Error("expected no plan to be saved on parse error")
	}
}

func TestGeneratePlanHandlerNoGateway(t *testing.T) {
	repo := &mockPlanRepo{}
	handler := GeneratePlanHandler("", "", repo, nil, nil)

	body := `{"age":30,"gender":"male","employment":"flexible","monthly_budget":3000}`
	req := httptest.NewRequest("POST", "/v1/plans/generate", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), middleware.ContextKeyUserID, "user-1")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d: %s", w.Code, w.Body.String())
	}
	respBody := w.Body.String()
	if !strings.Contains(respBody, "CONFIG_ERROR") {
		t.Errorf("expected CONFIG_ERROR, got %s", respBody)
	}
}

func TestParseLLMResponse(t *testing.T) {
	freeForm, structured, err := parseLLMResponse(mockLLMResponseBody)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(freeForm, "中等缴费基数") {
		t.Errorf("expected free form text about 中等缴费基数, got %s", freeForm)
	}
	if structured == nil {
		t.Fatal("expected structured response")
	}
	if structured.Summary != "推荐中等基数方案" {
		t.Errorf("expected summary '推荐中等基数方案', got %s", structured.Summary)
	}
	if len(structured.Schemes) != 1 {
		t.Errorf("expected 1 scheme, got %d", len(structured.Schemes))
	}
	if structured.Recommendation.RecommendedScheme != "方案A" {
		t.Errorf("expected recommended scheme '方案A', got %s", structured.Recommendation.RecommendedScheme)
	}
}

func TestParseLLMResponseMissingMarkers(t *testing.T) {
	_, _, err := parseLLMResponse("no markers here")
	if err == nil {
		t.Error("expected error for missing markers")
	}
}

func TestParseLLMResponseOnlyFreeForm(t *testing.T) {
	input := `===FREE_FORM_START===
some free text
===FREE_FORM_END===`
	freeForm, structured, err := parseLLMResponse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if freeForm != "some free text" {
		t.Errorf("expected 'some free text', got %s", freeForm)
	}
	if structured != nil {
		t.Error("expected nil structured when only free form present")
	}
}

var _ = fmt.Sprintf
