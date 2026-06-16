package handler

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/trigold786/94-AI-Insurance-Design/shared/auth"
	"github.com/trigold786/94-AI-Insurance-Design/shared/middleware"
)

type CodeStore interface {
	Set(ctx context.Context, key, value string, ttl time.Duration) error
	Get(ctx context.Context, key string) (string, error)
	Del(ctx context.Context, key string) error
}

func generateCode() string {
	n, _ := rand.Int(rand.Reader, big.NewInt(1000000))
	return fmt.Sprintf("%06d", n.Int64())
}

func SendSMSCodeHandler(codeStore CodeStore) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
		var req struct {
			Phone string `json:"phone"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondJSON(w, 400, map[string]interface{}{"code": "VALIDATION_ERROR", "message": "invalid JSON"})
			return
		}
		if len(req.Phone) != 11 {
			respondJSON(w, 400, map[string]interface{}{"code": "VALIDATION_ERROR", "message": "phone must be 11 digits"})
			return
		}

		rateKey := "sms_rate:" + req.Phone
		if rateCount, _ := codeStore.Get(r.Context(), rateKey); rateCount != "" {
			respondJSON(w, 429, map[string]interface{}{"code": "RATE_LIMITED", "message": "请稍后再试"})
			return
		}
		codeStore.Set(r.Context(), rateKey, "1", 60*time.Second)

		code := generateCode()
		key := "sms:" + req.Phone
		if err := codeStore.Set(r.Context(), key, code, 5*time.Minute); err != nil {
			respondError(w, err)
			return
		}
		codeStore.Set(r.Context(), "sms_attempts:"+req.Phone, "0", 5*time.Minute)

		if os.Getenv("LOG_SMS_CODE") == "true" {
			log.Printf("[sms] verification code for %s: %s", req.Phone, code)
		}
		respondJSON(w, 200, map[string]interface{}{"code": 0, "message": "验证码已发送"})
	})
}

func VerifySMSCodeHandler(codeStore CodeStore, jwtSecret string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
		var req struct {
			Phone string `json:"phone"`
			Code  string `json:"code"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondJSON(w, 400, map[string]interface{}{"code": "VALIDATION_ERROR", "message": "invalid JSON"})
			return
		}
		if req.Phone == "" || req.Code == "" {
			respondJSON(w, 400, map[string]interface{}{"code": "VALIDATION_ERROR", "message": "phone and code required"})
			return
		}

		key := "sms:" + req.Phone
		stored, err := codeStore.Get(r.Context(), key)
		if err != nil || stored == "" {
			respondJSON(w, 400, map[string]interface{}{"code": "VALIDATION_ERROR", "message": "验证码已过期，请重新发送"})
			return
		}

		attemptsKey := "sms_attempts:" + req.Phone
		attemptsStr, _ := codeStore.Get(r.Context(), attemptsKey)
		attempts := 0
		if attemptsStr != "" {
			fmt.Sscanf(attemptsStr, "%d", &attempts)
		}
		if attempts >= 5 {
			codeStore.Del(r.Context(), key)
			respondJSON(w, 400, map[string]interface{}{"code": "VALIDATION_ERROR", "message": "尝试次数过多，验证码已失效"})
			return
		}
		codeStore.Set(r.Context(), attemptsKey, fmt.Sprintf("%d", attempts+1), 5*time.Minute)

		if stored != req.Code {
			respondJSON(w, 400, map[string]interface{}{"code": "VALIDATION_ERROR", "message": "验证码错误"})
			return
		}
		codeStore.Del(r.Context(), key)

		userID := "u_" + req.Phone
		token, err := auth.GenerateToken(jwtSecret, userID, 24*7*time.Hour)
		if err != nil {
			respondJSON(w, 500, map[string]interface{}{"code": "INTERNAL_ERROR"})
			return
		}
		respondJSON(w, 200, map[string]interface{}{
			"code": 0,
			"data": map[string]string{
				"token":   token,
				"user_id": userID,
			},
		})
	})
}

func DeleteAccountHandler(codeStore CodeStore) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, _ := r.Context().Value(middleware.ContextKeyUserID).(string)
		if userID == "" {
			respondJSON(w, 401, map[string]interface{}{"code": "UNAUTHORIZED"})
			return
		}
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
		var req struct {
			Confirm string `json:"confirm"`
		}
		json.NewDecoder(r.Body).Decode(&req)
		if req.Confirm != "DELETE" {
			respondJSON(w, 400, map[string]interface{}{"code": "VALIDATION_ERROR", "message": "confirm field must be 'DELETE'"})
			return
		}
		key := "account_deleted:" + userID
		codeStore.Set(r.Context(), key, "1", 365*24*time.Hour)
		log.Printf("[auth] account deletion requested for user %s", userID)
		respondJSON(w, 200, map[string]interface{}{"code": 0, "message": "账号注销请求已受理，数据将在24小时内清除"})
	})
}

type MemoryCodeStore struct {
	mu   sync.RWMutex
	data map[string]memoryEntry
}

type memoryEntry struct {
	value   string
	expires time.Time
}

func NewMemoryCodeStore() *MemoryCodeStore {
	s := &MemoryCodeStore{data: make(map[string]memoryEntry)}
	go s.cleanup()
	return s
}

func (s *MemoryCodeStore) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		s.mu.Lock()
		now := time.Now()
		for k, v := range s.data {
			if now.After(v.expires) {
				delete(s.data, k)
			}
		}
		s.mu.Unlock()
	}
}

func (s *MemoryCodeStore) Set(_ context.Context, key, value string, ttl time.Duration) error {
	s.mu.Lock()
	s.data[key] = memoryEntry{value: value, expires: time.Now().Add(ttl)}
	s.mu.Unlock()
	return nil
}

func (s *MemoryCodeStore) Get(_ context.Context, key string) (string, error) {
	s.mu.RLock()
	entry, ok := s.data[key]
	s.mu.RUnlock()
	if !ok || time.Now().After(entry.expires) {
		return "", fmt.Errorf("not found")
	}
	return entry.value, nil
}

func (s *MemoryCodeStore) Del(_ context.Context, key string) error {
	s.mu.Lock()
	delete(s.data, key)
	s.mu.Unlock()
	return nil
}
