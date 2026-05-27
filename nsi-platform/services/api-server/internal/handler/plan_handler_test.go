package handler

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

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

type mockCalculator struct {
	resp *CalculateResponse
	err  error
}

func (m *mockCalculator) Calculate(ctx context.Context, req *CalculateRequest) (*CalculateResponse, error) {
	return m.resp, m.err
}

func TestGeneratePlanHandlerSuccess(t *testing.T) {
	calc := &mockCalculator{
		resp: &CalculateResponse{
			Schemes: []SchemeResult{
				{Name: "最低基数 (60%)", BaseSalary: 6000, MonthlyCost: 600, ProjectedPension: 2500},
			},
		},
	}
	repo := &mockPlanRepo{}
	handler := GeneratePlanHandler(calc, repo, nil, nil)

	body := `{"age":30,"gender":"male","employment":"flexible","contribution_years":10,"current_balance":50000,"monthly_budget":3000,"local_avg_salary":10000}`
	req := httptest.NewRequest("POST", "/v1/plans/generate", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "6000") {
		t.Errorf("expected base salary in response, got %s", w.Body.String())
	}
}

func TestGeneratePlanHandlerInvalidAge(t *testing.T) {
	handler := GeneratePlanHandler(&mockCalculator{}, nil, nil, nil)

	body := `{"age":-5,"gender":"male","monthly_budget":1000,"local_avg_salary":10000}`
	req := httptest.NewRequest("POST", "/v1/plans/generate", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestGeneratePlanHandlerInvalidGender(t *testing.T) {
	handler := GeneratePlanHandler(&mockCalculator{}, nil, nil, nil)

	body := `{"age":30,"gender":"other","monthly_budget":1000,"local_avg_salary":10000}`
	req := httptest.NewRequest("POST", "/v1/plans/generate", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestGeneratePlanHandlerInvalidBudget(t *testing.T) {
	handler := GeneratePlanHandler(&mockCalculator{}, nil, nil, nil)

	body := `{"age":30,"gender":"male","monthly_budget":0,"local_avg_salary":10000}`
	req := httptest.NewRequest("POST", "/v1/plans/generate", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestGeneratePlanHandlerInvalidJSON(t *testing.T) {
	handler := GeneratePlanHandler(nil, nil, nil, nil)

	req := httptest.NewRequest("POST", "/v1/plans/generate", strings.NewReader(`not json`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestGeneratePlanHandlerCalculatorError(t *testing.T) {
	calc := &mockCalculator{err: errCalculatorFailed}
	handler := GeneratePlanHandler(calc, &mockPlanRepo{}, nil, nil)

	body := `{"age":30,"gender":"male","employment":"flexible","contribution_years":10,"current_balance":50000,"monthly_budget":3000,"local_avg_salary":10000}`
	req := httptest.NewRequest("POST", "/v1/plans/generate", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

func TestGeneratePlanHandlerSavesPlan(t *testing.T) {
	calc := &mockCalculator{
		resp: &CalculateResponse{
			Schemes: []SchemeResult{{Name: "Test", BaseSalary: 5000, MonthlyCost: 500, ProjectedPension: 2000}},
		},
	}
	repo := &mockPlanRepo{}
	handler := GeneratePlanHandler(calc, repo, nil, nil)

	body := `{"age":35,"gender":"female","employment":"employed","contribution_years":15,"current_balance":100000,"monthly_budget":5000,"local_avg_salary":12000}`
	req := httptest.NewRequest("POST", "/v1/plans/generate", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if repo.savedPlan == nil {
		t.Fatal("expected plan to be saved")
	}
	if len(repo.savedPlan.RecommendedSchemes) != 1 {
		t.Errorf("expected 1 scheme, got %d", len(repo.savedPlan.RecommendedSchemes))
	}
}

var errCalculatorFailed = fmt.Errorf("calculator service unavailable")

func TestGeneratePlanHandlerCashflowPassthrough(t *testing.T) {
	calc := &mockCalculator{
		resp: &CalculateResponse{
			Schemes: []SchemeResult{
				{
					Name:            "缴费基数 6000",
					BaseSalary:      6000,
					MonthlyCost:     600,
					ProjectedPension: 2500,
					AfterTaxPension: 2300,
					Cashflow: []models.CashFlowItem{
						{Year: 1, Payment: 7200, Subsidy: 1200, Balance: 8400},
						{Year: 2, Payment: 7560, Subsidy: 1260, Balance: 17220},
					},
				},
			},
		},
	}
	repo := &mockPlanRepo{}
	handler := GeneratePlanHandler(calc, repo, nil, nil)

	body := `{"age":30,"gender":"male","employment":"flexible","contribution_years":10,"current_balance":50000,"monthly_budget":3000,"local_avg_salary":10000}`
	req := httptest.NewRequest("POST", "/v1/plans/generate", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	if repo.savedPlan == nil {
		t.Fatal("expected plan to be saved")
	}
	scheme := repo.savedPlan.RecommendedSchemes[0]
	if scheme.AfterTaxPension != 2300 {
		t.Errorf("expected afterTaxPension 2300, got %f", scheme.AfterTaxPension)
	}
	if len(scheme.Cashflow) != 2 {
		t.Fatalf("expected 2 cashflow items, got %d", len(scheme.Cashflow))
	}
	if scheme.Cashflow[0].Year != 1 || scheme.Cashflow[0].Payment != 7200 {
		t.Errorf("unexpected cashflow item: %+v", scheme.Cashflow[0])
	}
	respBody := w.Body.String()
	if !strings.Contains(respBody, "after_tax_pension") {
		t.Error("response should contain after_tax_pension field")
	}
	if !strings.Contains(respBody, "cashflow") {
		t.Error("response should contain cashflow field")
	}
}
