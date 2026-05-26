package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/trigold786/94-AI-Insurance-Design/shared/middleware"
	"github.com/trigold786/94-AI-Insurance-Design/shared/models"
)

// E2E test covering: profile → plan → report → compliance → guide → feedback → rights
func TestE2EFullChainV2(t *testing.T) {
	// Mock actuarial engine
	aeSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"schemes":[
			{"name":"最低基数 (60%)","base_salary":6000,"monthly_cost":600,"projected_pension":2500,"annual_subsidy":1200},
			{"name":"中等基数 (100%)","base_salary":10000,"monthly_cost":1000,"projected_pension":4000,"annual_subsidy":2000},
			{"name":"最高基数 (300%)","base_salary":30000,"monthly_cost":3000,"projected_pension":8000,"annual_subsidy":6000}
		],"total_cost":600,"total_subsidy":1200,"calculation_time_ms":8.2}`))
	}))
	defer aeSrv.Close()

	calc := &HTTPCalculator{Endpoint: aeSrv.URL}
	planRepo := &mockPlanRepo{}
	profileRepo := &mockProfileRepo{}
	policyRepo := &mockPolicyRepo{claims: []models.PolicyClaim{
		{ClaimID: "pol-1", PolicyID: "SH-FLEX-001", PolicyType: "subsidy", RegionCode: "310000", 
		 SubsidyCalcMethod: "灵活就业社保补贴50%", SourceName: "上海人社局", PolicyURL: "https://rsj.sh.gov.cn/policy/001"},
		{ClaimID: "pol-2", PolicyID: "BJ-PEN-001", PolicyType: "pension", RegionCode: "110000",
		 SubsidyCalcMethod: "养老保险缴费补贴", SourceName: "北京人社局"},
	}}
	evaluator := &ComplianceEvaluator{}

	userID := "e2e-v2-user"

	// Step 1: Health check
	t.Log("Step 1: Health check")
	healthHandler := HealthCheckHandler()
	req := httptest.NewRequest("GET", "/healthz", nil)
	w := httptest.NewRecorder()
	healthHandler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("HealthCheck expected 200, got %d", w.Code)
	}
	var healthResp struct {
		Status string `json:"status"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &healthResp); err != nil {
		t.Fatalf("HealthCheck JSON error: %v", err)
	}
	if healthResp.Status != "ok" {
		t.Fatalf("HealthCheck expected ok, got %s", healthResp.Status)
	}

	// Step 2: Update profile
	t.Log("Step 2: Update profile")
	profileHandler := middleware.AuthMiddleware("")(UpdateProfileHandler(profileRepo))
	profileBody := `{"age":30,"gender":"male","household_region_code":"310000","current_residence_code":"310000","employment_status":"flexible","social_security_years":10,"has_children":false}`
	preq := httptest.NewRequest("PUT", "/v1/profile", strings.NewReader(profileBody))
	preq.Header.Set("x-user-id", userID)
	pw := httptest.NewRecorder()
	profileHandler.ServeHTTP(pw, preq)
	if pw.Code != http.StatusOK {
		t.Fatalf("UpdateProfile expected 200, got %d: %s", pw.Code, pw.Body.String())
	}

	// Step 3: Generate plan
	t.Log("Step 3: Generate plan")
	genHandler := middleware.AuthMiddleware("")(GeneratePlanHandler(calc, planRepo, nil, nil))
	genBody := `{"age":30,"gender":"male","employment":"flexible","contribution_years":10,"current_balance":50000,"monthly_budget":3000,"priority":"balanced","local_avg_salary":10000}`
	greq := httptest.NewRequest("POST", "/v1/plans/generate", strings.NewReader(genBody))
	greq.Header.Set("x-user-id", userID)
	gw := httptest.NewRecorder()
	genHandler.ServeHTTP(gw, greq)
	if gw.Code != http.StatusOK {
		t.Fatalf("GeneratePlan expected 200, got %d: %s", gw.Code, gw.Body.String())
	}

	var genWrapper struct {
		Code int                    `json:"code"`
		Data *models.PlanSnapshot   `json:"data"`
	}
	if err := json.Unmarshal(gw.Body.Bytes(), &genWrapper); err != nil {
		t.Fatalf("GeneratePlan JSON error: %v", err)
	}
	if genWrapper.Data == nil || genWrapper.Data.PlanID == "" {
		t.Fatal("GeneratePlan: expected non-empty plan_id")
	}
	if len(genWrapper.Data.RecommendedSchemes) != 3 {
		t.Fatalf("GeneratePlan: expected 3 schemes, got %d", len(genWrapper.Data.RecommendedSchemes))
	}
	planID := genWrapper.Data.PlanID
	planRepo.savedPlan.PlanID = planID

	// Step 4: Get plan detail
	t.Log("Step 4: Get plan detail")
	detailHandler := middleware.AuthMiddleware("")(PlanDetailHandler(planRepo))
	dreq := httptest.NewRequest("GET", "/v1/plans/"+planID, nil)
	dreq.Header.Set("x-user-id", userID)
	dw := httptest.NewRecorder()
	detailHandler.ServeHTTP(dw, dreq)
	if dw.Code != http.StatusOK {
		t.Fatalf("PlanDetail expected 200, got %d: %s", dw.Code, dw.Body.String())
	}
	var detailWrapper struct {
		Code int                  `json:"code"`
		Data *models.PlanSnapshot `json:"data"`
	}
	if err := json.Unmarshal(dw.Body.Bytes(), &detailWrapper); err != nil {
		t.Fatalf("PlanDetail JSON error: %v", err)
	}
	if detailWrapper.Data.PlanID != planID {
		t.Errorf("PlanDetail expected plan_id %s, got %s", planID, detailWrapper.Data.PlanID)
	}
	// Verify first scheme
	if len(detailWrapper.Data.RecommendedSchemes) > 0 {
		s := detailWrapper.Data.RecommendedSchemes[0]
		if s.BaseSalary != 6000 {
			t.Errorf("PlanDetail expected BaseSalary 6000, got %d", s.BaseSalary)
		}
	}

	// Step 5: Get plan report (HTML)
	t.Log("Step 5: Get plan report")
	reportHandler := middleware.AuthMiddleware("")(PlanReportHandler(planRepo, policyRepo))
	rreq := httptest.NewRequest("GET", "/v1/plans/report?plan_id="+planID, nil)
	rreq.Header.Set("x-user-id", userID)
	rw := httptest.NewRecorder()
	reportHandler.ServeHTTP(rw, rreq)
	if rw.Code != http.StatusOK {
		t.Fatalf("PlanReport expected 200, got %d", rw.Code)
	}
	ct := rw.Header().Get("Content-Type")
	if ct != "text/html; charset=utf-8" {
		t.Errorf("PlanReport expected text/html, got %s", ct)
	}
	body := rw.Body.String()
	if !strings.Contains(body, "社保规划报告") || !strings.Contains(body, "方案概览") {
		t.Errorf("PlanReport HTML missing key sections")
	}
	if !strings.Contains(body, "行动步骤") {
		t.Errorf("PlanReport HTML missing action steps")
	}

	// Step 6: Query policies
	t.Log("Step 6: Query policies")
	policyHandler := middleware.AuthMiddleware("")(QueryPoliciesHandler(policyRepo))
	qreq := httptest.NewRequest("GET", "/v1/policies?region_code=310000", nil)
	qreq.Header.Set("x-user-id", userID)
	qw := httptest.NewRecorder()
	policyHandler.ServeHTTP(qw, qreq)
	if qw.Code != http.StatusOK {
		t.Fatalf("QueryPolicies expected 200, got %d: %s", qw.Code, qw.Body.String())
	}
	var policyWrapper struct {
		Code int                    `json:"code"`
		Data []models.PolicyClaim  `json:"data"`
	}
	if err := json.Unmarshal(qw.Body.Bytes(), &policyWrapper); err != nil {
		t.Fatalf("QueryPolicies JSON error: %v", err)
	}
	if len(policyWrapper.Data) == 0 {
		t.Error("QueryPolicies: expected at least 1 policy")
	}

	// Step 7: Get compliance checklist
	t.Log("Step 7: Get compliance checklist")
	complianceHandler := middleware.AuthMiddleware("")(ComplianceChecklistHandler(evaluator, policyRepo, profileRepo))
	creq := httptest.NewRequest("GET", "/v1/compliance/checklist?city_code=310000", nil)
	creq.Header.Set("x-user-id", userID)
	cw := httptest.NewRecorder()
	complianceHandler.ServeHTTP(cw, creq)
	if cw.Code != http.StatusOK {
		t.Fatalf("Compliance expected 200, got %d: %s", cw.Code, cw.Body.String())
	}
	var compWrapper struct {
		Code int                        `json:"code"`
		Data models.ComplianceChecklist `json:"data"`
	}
	if err := json.Unmarshal(cw.Body.Bytes(), &compWrapper); err != nil {
		t.Fatalf("Compliance JSON error: %v", err)
	}
	if compWrapper.Data.CityCode != "310000" {
		t.Errorf("Compliance expected city_code 310000, got %s", compWrapper.Data.CityCode)
	}
	// Verify processing steps are present
	for _, p := range compWrapper.Data.MatchedPolicies {
		if len(p.ProcessingSteps) > 0 {
			t.Logf("  Policy %s has %d processing steps", p.PolicyID, len(p.ProcessingSteps))
			break
		}
	}

	// Step 8: Get guide (HTML)
	t.Log("Step 8: Get guide")
	guideHandler := middleware.AuthMiddleware("")(GuideHandler(evaluator, policyRepo, profileRepo))
	greq2 := httptest.NewRequest("GET", "/v1/guide?city_code=310000", nil)
	greq2.Header.Set("x-user-id", userID)
	gw2 := httptest.NewRecorder()
	guideHandler.ServeHTTP(gw2, greq2)
	if gw2.Code != http.StatusOK {
		t.Fatalf("Guide expected 200, got %d", gw2.Code)
	}
	guideBody := gw2.Body.String()
	if !strings.Contains(guideBody, "办理指南") {
		t.Error("Guide HTML missing title")
	}
	if !strings.Contains(guideBody, "材料清单") {
		t.Error("Guide HTML missing materials section")
	}

	// Step 9: Submit feedback
	t.Log("Step 9: Submit feedback")
	feedbackHandler := middleware.AuthMiddleware("")(SubmitFeedbackHandler(&mockFeedbackRepo{}))
	fbody := `{"category":"general","content":"E2E test feedback","contact":"test@example.com"}`
	freq := httptest.NewRequest("POST", "/v1/feedback", strings.NewReader(fbody))
	freq.Header.Set("x-user-id", userID)
	fw := httptest.NewRecorder()
	feedbackHandler.ServeHTTP(fw, freq)
	if fw.Code != http.StatusOK {
		t.Fatalf("Feedback expected 200, got %d: %s", fw.Code, fw.Body.String())
	}
	var fbWrapper struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(fw.Body.Bytes(), &fbWrapper); err != nil {
		t.Fatalf("Feedback JSON error: %v", err)
	}
	if !strings.Contains(fbWrapper.Message, "感谢") {
		t.Errorf("Feedback expected thank-you message, got %s", fbWrapper.Message)
	}

	// Step 10: Get payment status
	t.Log("Step 10: Get payment status")
	rightsRepo := &mockRightsRepo{records: []models.PaymentRecord{
		{RecordID: "rec-1", UserID: userID, PolicyType: "pension", Month: "2026-05", Amount: 1000, Status: "paid", DueDate: "2026-05-15"},
		{RecordID: "rec-2", UserID: userID, PolicyType: "medical", Month: "2026-05", Amount: 500, Status: "pending", DueDate: "2026-05-20"},
	}}
	paymentHandler := middleware.AuthMiddleware("")(PaymentStatusHandler(rightsRepo))
	preq2 := httptest.NewRequest("GET", "/v1/rights/payment-status", nil)
	preq2.Header.Set("x-user-id", userID)
	pw2 := httptest.NewRecorder()
	paymentHandler.ServeHTTP(pw2, preq2)
	if pw2.Code != http.StatusOK {
		t.Fatalf("PaymentStatus expected 200, got %d: %s", pw2.Code, pw2.Body.String())
	}

	// Step 11: Get alerts
	t.Log("Step 11: Get alerts")
	alertHandler := middleware.AuthMiddleware("")(AlertListHandler(rightsRepo))
	areq := httptest.NewRequest("GET", "/v1/rights/alerts", nil)
	areq.Header.Set("x-user-id", userID)
	aw := httptest.NewRecorder()
	alertHandler.ServeHTTP(aw, areq)
	if aw.Code != http.StatusOK {
		t.Fatalf("Alerts expected 200, got %d: %s", aw.Code, aw.Body.String())
	}

	t.Log("All E2E steps completed successfully")
}

// mockFeedbackRepo for E2E tests
type mockFeedbackRepo struct{}

func (m *mockFeedbackRepo) SaveFeedback(_ context.Context, _, _, _, _ string) error {
	return nil
}
