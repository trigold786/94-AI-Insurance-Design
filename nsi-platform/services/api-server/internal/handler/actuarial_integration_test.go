package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/trigold786/94-AI-Insurance-Design/shared/middleware"
	"github.com/trigold786/94-AI-Insurance-Design/shared/models"
)

func TestHTTPCalculatorToActuarialEngine(t *testing.T) {
	aeSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		var req CalculateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		resp := CalculateResponse{
			Schemes: []SchemeResult{
				{Name: "最低基数 (60%)", BaseSalary: int(float64(req.LocalAvgSalary) * 0.6), MonthlyCost: 600, ProjectedPension: 2500},
				{Name: "中等基数 (100%)", BaseSalary: int(req.LocalAvgSalary), MonthlyCost: 1000, ProjectedPension: 4000},
			},
			CalculationTimeMs: 12.5,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer aeSrv.Close()

	calc := &HTTPCalculator{Endpoint: aeSrv.URL}

	resp, err := calc.Calculate(context.Background(), &CalculateRequest{
		Age:               30,
		Gender:            "male",
		Employment:        "flexible",
		ContributionYears: 10,
		CurrentBalance:    50000,
		MonthlyBudget:     3000,
		Priority:          "balanced",
		LocalAvgSalary:    10000,
	})
	if err != nil {
		t.Fatalf("Calculate() returned error: %v", err)
	}

	if len(resp.Schemes) != 2 {
		t.Fatalf("expected 2 schemes, got %d", len(resp.Schemes))
	}

	if resp.Schemes[0].Name != "最低基数 (60%)" {
		t.Errorf("expected '最低基数 (60 percent)', got '%s'", resp.Schemes[0].Name)
	}

	if resp.Schemes[1].BaseSalary != 10000 {
		t.Errorf("expected BaseSalary 10000, got %d", resp.Schemes[1].BaseSalary)
	}

	if resp.CalculationTimeMs <= 0 {
		t.Errorf("expected positive CalculationTimeMs, got %f", resp.CalculationTimeMs)
	}
}

func TestHTTPCalculatorServerError(t *testing.T) {
	aeSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer aeSrv.Close()

	calc := &HTTPCalculator{Endpoint: aeSrv.URL}
	_, err := calc.Calculate(context.Background(), &CalculateRequest{
		Age: 30, MonthlyBudget: 3000, LocalAvgSalary: 10000,
	})
	if err == nil {
		t.Fatal("expected error for 500 response, got nil")
	}
}

func TestHTTPCalculatorBadJSON(t *testing.T) {
	aeSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{invalid`))
	}))
	defer aeSrv.Close()

	calc := &HTTPCalculator{Endpoint: aeSrv.URL}
	_, err := calc.Calculate(context.Background(), &CalculateRequest{
		Age: 30, MonthlyBudget: 3000, LocalAvgSalary: 10000,
	})
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}

func TestHTTPCalculatorGeneratePlanFullFlow(t *testing.T) {
	aeSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"schemes":[{"name":"方案A","base_salary":6000,"monthly_cost":600,"projected_pension":2500}],"calculation_time_ms":8.2}`))
	}))
	defer aeSrv.Close()

	calc := &HTTPCalculator{Endpoint: aeSrv.URL}
	repo := &mockPlanRepo{}

	handler := middleware.AuthMiddleware("")(GeneratePlanHandler(calc, repo, nil, nil))

	body := bytes.NewBufferString(`{"age":30,"gender":"male","employment":"flexible","contribution_years":10,"current_balance":50000,"monthly_budget":3000,"priority":"balanced","local_avg_salary":10000}`)
	req := httptest.NewRequest("POST", "/v1/plans/generate", body)
	req.Header.Set("x-user-id", "user-int-1")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	if repo.savedPlan == nil {
		t.Fatal("expected plan to be saved")
	}
	if repo.savedPlan.UserID != "user-int-1" {
		t.Errorf("expected UserID user-int-1, got %s", repo.savedPlan.UserID)
	}
	if len(repo.savedPlan.RecommendedSchemes) != 1 {
		t.Errorf("expected 1 scheme, got %d", len(repo.savedPlan.RecommendedSchemes))
	}
}

func TestE2EFullChain(t *testing.T) {
	// Actuarial-engine mock
	aeSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"schemes":[{"name":"方案A","base_salary":6000,"monthly_cost":600,"projected_pension":2500,"annual_subsidy":1200}],"total_cost":600,"total_subsidy":1200,"calculation_time_ms":8.2}`))
	}))
	defer aeSrv.Close()

	calc := &HTTPCalculator{Endpoint: aeSrv.URL}
	repo := &mockPlanRepo{}

	userID := "e2e-user-1"

	// Step 1: UpdateProfile
	profileHandler := middleware.AuthMiddleware("")(UpdateProfileHandler(&mockProfileRepo{}))
	profileBody := `{"age":30,"gender":"male","household_region_code":"310000","current_residence_code":"310000","employment_status":"flexible","social_security_years":10,"has_children":false}`
	preq := httptest.NewRequest("PUT", "/v1/profile", strings.NewReader(profileBody))
	preq.Header.Set("x-user-id", userID)
	pw := httptest.NewRecorder()
	profileHandler.ServeHTTP(pw, preq)
	if pw.Code != http.StatusOK {
		t.Fatalf("UpdateProfile expected 200, got %d: %s", pw.Code, pw.Body.String())
	}

	// Step 2: GeneratePlan
	genHandler := middleware.AuthMiddleware("")(GeneratePlanHandler(calc, repo, nil, nil))
	genBody := `{"age":30,"gender":"male","employment":"flexible","contribution_years":10,"current_balance":50000,"monthly_budget":3000,"priority":"balanced","local_avg_salary":10000}`
	greq := httptest.NewRequest("POST", "/v1/plans/generate", strings.NewReader(genBody))
	greq.Header.Set("x-user-id", userID)
	gw := httptest.NewRecorder()
	genHandler.ServeHTTP(gw, greq)
	if gw.Code != http.StatusOK {
		t.Fatalf("GeneratePlan expected 200, got %d: %s", gw.Code, gw.Body.String())
	}

	var wrapper struct {
		Code int              `json:"code"`
		Data *models.PlanSnapshot `json:"data"`
	}
	if err := json.Unmarshal(gw.Body.Bytes(), &wrapper); err != nil {
		t.Fatalf("failed to decode generate response: %v", err)
	}
	if wrapper.Data == nil || wrapper.Data.PlanID == "" {
		t.Fatal("expected non-empty plan_id in generate response")
	}
	if len(wrapper.Data.RecommendedSchemes) == 0 {
		t.Fatal("expected at least 1 scheme in generate response")
	}

	planID := wrapper.Data.PlanID
	repo.savedPlan.PlanID = planID

	// Step 3: GetPlanDetail
	detailHandler := middleware.AuthMiddleware("")(PlanDetailHandler(repo))
	dreq := httptest.NewRequest("GET", "/v1/plans/"+planID, nil)
	dreq.Header.Set("x-user-id", userID)
	dw := httptest.NewRecorder()
	detailHandler.ServeHTTP(dw, dreq)
	if dw.Code != http.StatusOK {
		t.Fatalf("GetPlanDetail expected 200, got %d: %s", dw.Code, dw.Body.String())
	}

	var detailWrapper struct {
		Code int              `json:"code"`
		Data *models.PlanSnapshot `json:"data"`
	}
	if err := json.Unmarshal(dw.Body.Bytes(), &detailWrapper); err != nil {
		t.Fatalf("failed to decode detail response: %v", err)
	}
	if detailWrapper.Data == nil || detailWrapper.Data.PlanID != planID {
		t.Errorf("expected plan_id %s, got %v", planID, detailWrapper.Data)
	}
	if len(detailWrapper.Data.RecommendedSchemes) != 1 {
		t.Errorf("expected 1 scheme, got %d", len(detailWrapper.Data.RecommendedSchemes))
	}
	if detailWrapper.Data.RecommendedSchemes[0].BaseSalary != 6000 {
		t.Errorf("expected BaseSalary 6000, got %d", detailWrapper.Data.RecommendedSchemes[0].BaseSalary)
	}

	// Step 4: GetPlanDetail with wrong user (should 404)
	dreq2 := httptest.NewRequest("GET", "/v1/plans/"+planID, nil)
	dreq2.Header.Set("x-user-id", "wrong-user")
	dw2 := httptest.NewRecorder()
	detailHandler.ServeHTTP(dw2, dreq2)
	if dw2.Code != http.StatusNotFound {
		t.Errorf("expected 404 for wrong user, got %d", dw2.Code)
	}
}
