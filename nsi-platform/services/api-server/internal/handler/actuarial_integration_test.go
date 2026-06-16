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

func TestLLMGatewayChatSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/v1/chat" {
			w.WriteHeader(404)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(LLMChatResponse{
			Code:    200,
			Content: mockLLMResponseBody,
		})
	}))
	defer srv.Close()

	client := &LLMGatewayClient{URL: srv.URL}
	resp, err := client.Chat(context.Background(), "system", "user content")
	if err != nil {
		t.Fatalf("Chat() returned error: %v", err)
	}
	if resp.Content == "" {
		t.Error("expected non-empty content")
	}
}

func TestLLMGatewayChatServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(LLMChatResponse{Code: 500, Content: "internal error"})
	}))
	defer srv.Close()

	client := &LLMGatewayClient{URL: srv.URL}
	_, err := client.Chat(context.Background(), "system", "user")
	if err == nil {
		t.Fatal("expected error for 500 response, got nil")
	}
}

func TestLLMGatewayChatBadJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{invalid`))
	}))
	defer srv.Close()

	client := &LLMGatewayClient{URL: srv.URL}
	_, err := client.Chat(context.Background(), "system", "user")
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}

func TestLLMGatewayGeneratePlanFullFlow(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(LLMChatResponse{Code: 200, Content: mockLLMResponseBody})
	}))
	defer srv.Close()

	repo := &mockPlanRepo{}
	handler := middleware.AuthMiddleware(testJWTSecret)(GeneratePlanHandler(srv.URL, "", repo, nil, nil))

	req := httptest.NewRequest("POST", "/v1/plans/generate", strings.NewReader(`{"age":30,"gender":"male","employment":"flexible","monthly_budget":3000}`))
	setAuth(req, "user-int-1")
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
	if len(repo.savedPlan.StructuredSchemes) != 1 {
		t.Errorf("expected 1 scheme, got %d", len(repo.savedPlan.StructuredSchemes))
	}
}

func TestE2EFullChain(t *testing.T) {
	llmSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(LLMChatResponse{Code: 200, Content: mockLLMResponseBody})
	}))
	defer llmSrv.Close()

	repo := &mockPlanRepo{}
	userID := "e2e-user-1"

	profileHandler := middleware.AuthMiddleware(testJWTSecret)(UpdateProfileHandler(&mockProfileRepo{}))
	profileBody := `{"age":30,"gender":"male","household_region_code":"310000","current_residence_code":"310000","employment_status":"flexible","social_security_years":10,"has_children":false}`
	preq := httptest.NewRequest("PUT", "/v1/profile", strings.NewReader(profileBody))
	setAuth(preq, userID)
	pw := httptest.NewRecorder()
	profileHandler.ServeHTTP(pw, preq)
	if pw.Code != http.StatusOK {
		t.Fatalf("UpdateProfile expected 200, got %d: %s", pw.Code, pw.Body.String())
	}

	genHandler := middleware.AuthMiddleware(testJWTSecret)(GeneratePlanHandler(llmSrv.URL, "", repo, nil, nil))
	genBody := `{"age":30,"gender":"male","employment":"flexible","monthly_budget":3000}`
	greq := httptest.NewRequest("POST", "/v1/plans/generate", strings.NewReader(genBody))
	setAuth(greq, userID)
	gw := httptest.NewRecorder()
	genHandler.ServeHTTP(gw, greq)
	if gw.Code != http.StatusOK {
		t.Fatalf("GeneratePlan expected 200, got %d: %s", gw.Code, gw.Body.String())
	}

	var wrapper struct {
		Code int                `json:"code"`
		Data *models.PlanSnapshot `json:"data"`
	}
	if err := json.Unmarshal(gw.Body.Bytes(), &wrapper); err != nil {
		t.Fatalf("failed to decode generate response: %v", err)
	}
	if wrapper.Data == nil || wrapper.Data.PlanID == "" {
		t.Fatal("expected non-empty plan_id in generate response")
	}
	if len(wrapper.Data.StructuredSchemes) == 0 {
		t.Fatal("expected at least 1 scheme in generate response")
	}

	planID := wrapper.Data.PlanID
	repo.savedPlan.PlanID = planID

	detailHandler := middleware.AuthMiddleware(testJWTSecret)(PlanDetailHandler(repo))
	dreq := httptest.NewRequest("GET", "/v1/plans/"+planID, nil)
	setAuth(dreq, userID)
	dw := httptest.NewRecorder()
	detailHandler.ServeHTTP(dw, dreq)
	if dw.Code != http.StatusOK {
		t.Fatalf("GetPlanDetail expected 200, got %d: %s", dw.Code, dw.Body.String())
	}

	var detailWrapper struct {
		Code int                `json:"code"`
		Data *models.PlanSnapshot `json:"data"`
	}
	if err := json.Unmarshal(dw.Body.Bytes(), &detailWrapper); err != nil {
		t.Fatalf("failed to decode detail response: %v", err)
	}
	if detailWrapper.Data == nil || detailWrapper.Data.PlanID != planID {
		t.Errorf("expected plan_id %s, got %v", planID, detailWrapper.Data)
	}
	if len(detailWrapper.Data.StructuredSchemes) != 1 {
		t.Errorf("expected 1 scheme, got %d", len(detailWrapper.Data.StructuredSchemes))
	}
	if detailWrapper.Data.StructuredSchemes[0].ContributionBase != 6000 {
		t.Errorf("expected ContributionBase 6000, got %f", detailWrapper.Data.StructuredSchemes[0].ContributionBase)
	}

	dreq2 := httptest.NewRequest("GET", "/v1/plans/"+planID, nil)
	setAuth(dreq2, "wrong-user")
	dw2 := httptest.NewRecorder()
	detailHandler.ServeHTTP(dw2, dreq2)
	if dw2.Code != http.StatusNotFound {
		t.Errorf("expected 404 for wrong user, got %d", dw2.Code)
	}
}
