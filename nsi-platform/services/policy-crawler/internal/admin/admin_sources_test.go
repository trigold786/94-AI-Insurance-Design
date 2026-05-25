package admin

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSourceCreateHandler_MissingFields(t *testing.T) {
	handler := SourceCreateHandler(nil)
	body := `{"source_id":"TEST-1"}`
	req := httptest.NewRequest("POST", "/admin/sources/create", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestSourceCreateHandler_InvalidJSON(t *testing.T) {
	handler := SourceCreateHandler(nil)
	req := httptest.NewRequest("POST", "/admin/sources/create", strings.NewReader("not json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestSourceDeleteHandler_MissingID(t *testing.T) {
	handler := SourceDeleteHandler(nil)
	body := `{}`
	req := httptest.NewRequest("POST", "/admin/sources/delete", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestSourceCrawlTriggerHandler_MissingID(t *testing.T) {
	handler := SourceCrawlTriggerHandler(nil)
	body := `{}`
	req := httptest.NewRequest("POST", "/admin/sources/crawl", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestSourceCrawlTriggerHandler_Success(t *testing.T) {
	var triggered string
	_ = triggered
	mock := &mockCrawlTrigger{fn: func(id string) { triggered = id }}
	handler := SourceCrawlTriggerHandler(mock)
	body := `{"source_id":"TEST-SRC"}`
	req := httptest.NewRequest("POST", "/admin/sources/crawl", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["code"] != float64(0) {
		t.Errorf("expected code 0, got %v", resp["code"])
	}
}

func TestRSSTestHandler_MissingURL(t *testing.T) {
	handler := RSSTestHandler()
	body := `{}`
	req := httptest.NewRequest("POST", "/admin/sources/test-rss", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestRSSTestHandler_InvalidURL(t *testing.T) {
	handler := RSSTestHandler()
	body := `{"url":"http://127.0.0.1:1/nonexistent"}`
	req := httptest.NewRequest("POST", "/admin/sources/test-rss", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code == http.StatusOK {
		t.Errorf("expected non-200 for invalid URL, got %d", w.Code)
	}
}

func TestSourceUpdateHandler_FullFields(t *testing.T) {
	handler := SourceUpdateHandler(nil)
	body := `{"source_id":"TEST","source_name":"new name","source_url":"http://x","source_level":"LOW","crawl_type":"rss","region_code":"110000","interval_sec":3600}`
	req := httptest.NewRequest("POST", "/admin/sources/update", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

type mockCrawlTrigger struct {
	fn func(string)
}

func (m *mockCrawlTrigger) CrawlSource(id string) {
	if m.fn != nil {
		m.fn(id)
	}
}
