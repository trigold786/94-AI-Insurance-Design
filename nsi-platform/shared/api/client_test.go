package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/trigold786/94-AI-Insurance-Design/shared/models"
)

func TestNewClientInvalidURL(t *testing.T) {
	_, err := NewClient("http://[invalid:address]:port", "user-1")
	if err == nil {
		t.Fatal("expected error for invalid URL, got nil")
	}
}

func TestNewClientEmptyURL(t *testing.T) {
	_, err := NewClient("", "user-1")
	if err == nil {
		t.Fatal("expected error for empty URL, got nil")
	}
}

func TestNewClientSuccess(t *testing.T) {
	client, err := NewClient("http://localhost:30001", "user-1")
	if err != nil {
		t.Fatalf("NewClient() returned error: %v", err)
	}
	if client == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestGetProfile(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("x-user-id") != "test-user" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		if r.URL.Path != "/v1/profile" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"code":0,"data":{"user_id":"test-user","age":30,"gender":"male"}}`))
	}))
	defer srv.Close()

	client, _ := NewClient(srv.URL, "test-user")
	profile, err := client.GetProfile(context.Background())
	if err != nil {
		t.Fatalf("GetProfile() returned error: %v", err)
	}
	if profile.UserID != "test-user" {
		t.Errorf("expected test-user, got %s", profile.UserID)
	}
	if profile.Age != 30 {
		t.Errorf("expected age 30, got %d", profile.Age)
	}
}

func TestGetProfileUnauthorized(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"code":"UNAUTHORIZED","message":"missing x-user-id"}`))
	}))
	defer srv.Close()

	client, _ := NewClient(srv.URL, "wrong-user")
	_, err := client.GetProfile(context.Background())
	if err == nil {
		t.Fatal("expected error for unauthorized, got nil")
	}
}

func TestUpdateProfile(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"code":0,"data":{"user_id":"test-user","age":35}}`))
	}))
	defer srv.Close()

	client, _ := NewClient(srv.URL, "test-user")
	err := client.UpdateProfile(context.Background(), &models.UserProfile{Age: 35})
	if err != nil {
		t.Fatalf("UpdateProfile() returned error: %v", err)
	}
}

func TestQueryPolicies(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.RawQuery, "region_code=310000") {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"code":0,"data":[{"claim_id":"CLM-001","policy_id":"SH-001"}]}`))
	}))
	defer srv.Close()

	client, _ := NewClient(srv.URL, "test-user")
	claims, err := client.QueryPolicies(context.Background(), "310000", "", "")
	if err != nil {
		t.Fatalf("QueryPolicies() returned error: %v", err)
	}
	if len(claims) != 1 {
		t.Errorf("expected 1 claim, got %d", len(claims))
	}
	if claims[0].PolicyID != "SH-001" {
		t.Errorf("expected SH-001, got %s", claims[0].PolicyID)
	}
}

func TestGeneratePlan(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"code":0,"data":{"plan_id":"plan-1","recommended_schemes":[{"name":"Test","base_salary":6000}]}}`))
	}))
	defer srv.Close()

	client, _ := NewClient(srv.URL, "test-user")
	plan, err := client.GeneratePlan(context.Background(), `{"age":30}`)
	if err != nil {
		t.Fatalf("GeneratePlan() returned error: %v", err)
	}
	if plan.PlanID != "plan-1" {
		t.Errorf("expected plan-1, got %s", plan.PlanID)
	}
}

func TestClientSetsHeader(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("x-user-id") != "custom-user" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"code":0,"data":{}}`))
	}))
	defer srv.Close()

	client, _ := NewClient(srv.URL, "custom-user")
	_, err := client.GetProfile(context.Background())
	if err != nil {
		t.Fatalf("GetProfile() returned error: %v", err)
	}
}
