package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/trigold786/94-AI-Insurance-Design/shared/middleware"
)

type FeedbackRepo interface {
	SaveFeedback(ctx context.Context, userID, category, content, contact string) error
}

func SubmitFeedbackHandler(repo FeedbackRepo) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, _ := r.Context().Value(middleware.ContextKeyUserID).(string)
		if userID == "" {
			respondJSON(w, http.StatusUnauthorized, map[string]interface{}{"code": "UNAUTHORIZED", "message": "missing user"})
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
		var req struct {
			Category string `json:"category"`
			Content  string `json:"content"`
			Contact  string `json:"contact"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondJSON(w, http.StatusBadRequest, map[string]interface{}{"code": "VALIDATION_ERROR", "message": "invalid JSON"})
			return
		}
		if req.Content == "" {
			respondJSON(w, http.StatusBadRequest, map[string]interface{}{"code": "VALIDATION_ERROR", "message": "content required"})
			return
		}
		if len(req.Content) > 10000 {
			respondJSON(w, http.StatusBadRequest, map[string]interface{}{"code": "BAD_REQUEST", "message": "content too long (max 10000 chars)"})
			return
		}
		if req.Category == "" {
			req.Category = "general"
		}

		if err := repo.SaveFeedback(r.Context(), userID, req.Category, req.Content, req.Contact); err != nil {
			respondError(w, err)
			return
		}

		respondJSON(w, http.StatusOK, map[string]interface{}{"code": 0, "message": "反馈已提交，感谢您的意见！"})
	})
}
