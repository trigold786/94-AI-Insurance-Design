package handler

import (
	"net/http"
	"time"

	"github.com/trigold786/94-AI-Insurance-Design/shared/auth"
)

const testJWTSecret = "test-secret-123"

func setAuth(req *http.Request, userID string) *http.Request {
	token, _ := auth.GenerateToken(testJWTSecret, userID, time.Hour)
	req.Header.Set("Authorization", "Bearer "+token)
	return req
}
