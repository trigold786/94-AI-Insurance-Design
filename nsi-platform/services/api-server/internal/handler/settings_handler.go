package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/trigold786/94-AI-Insurance-Design/shared/middleware"
	"github.com/trigold786/94-AI-Insurance-Design/shared/models"
)

type SettingsStore interface {
	GetSettings(ctx context.Context, userID string) (*models.SettingsData, error)
	SaveSettings(ctx context.Context, userID string, s *models.SettingsData) error
	DeleteUserData(ctx context.Context, userID string) error
}

func GetSettingsHandler(store SettingsStore) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, _ := r.Context().Value(middleware.ContextKeyUserID).(string)
		s, err := store.GetSettings(r.Context(), userID)
		if err != nil || s == nil {
			s = &models.SettingsData{FontScale: "medium", DefaultTab: "profile", NotificationsOn: true}
		}
		respondJSON(w, 200, map[string]interface{}{"code": 0, "data": s})
	})
}

func SaveSettingsHandler(store SettingsStore) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, _ := r.Context().Value(middleware.ContextKeyUserID).(string)
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
		var s models.SettingsData
		if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
			respondJSON(w, 400, map[string]interface{}{"code": "VALIDATION_ERROR"})
			return
		}
		validFont := map[string]bool{"small": true, "medium": true, "large": true}
		if !validFont[s.FontScale] {
			s.FontScale = "medium"
		}
		validTabs := map[string]bool{"profile": true, "plan": true, "sandbox": true, "compliance": true, "rights": true}
		if !validTabs[s.DefaultTab] {
			s.DefaultTab = "profile"
		}
		if err := store.SaveSettings(r.Context(), userID, &s); err != nil {
			respondError(w, err)
			return
		}
		respondJSON(w, 200, map[string]interface{}{"code": 0, "data": s})
	})
}

func DeleteAccountHandlerV2(store SettingsStore) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, _ := r.Context().Value(middleware.ContextKeyUserID).(string)
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
		var req struct {
			Confirm string `json:"confirm"`
		}
		json.NewDecoder(r.Body).Decode(&req)
		if strings.ToUpper(req.Confirm) != "DELETE" {
			respondJSON(w, 400, map[string]interface{}{"code": "VALIDATION_ERROR", "message": "confirm must be 'DELETE'"})
			return
		}
		if err := store.DeleteUserData(r.Context(), userID); err != nil {
			respondError(w, err)
			return
		}
		respondJSON(w, 200, map[string]interface{}{"code": 0, "message": "账号已注销，所有数据已删除"})
	})
}
