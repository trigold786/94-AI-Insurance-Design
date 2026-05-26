package middleware

import (
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/trigold786/94-AI-Insurance-Design/shared/auth"
)

type contextKey string

const ContextKeyUserID contextKey = "user_id"

type Middleware func(http.Handler) http.Handler

func Chain(middlewares ...Middleware) Middleware {
	return func(next http.Handler) http.Handler {
		for i := len(middlewares) - 1; i >= 0; i-- {
			next = middlewares[i](next)
		}
		return next
	}
}

func AuthMiddleware(jwtSecret string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if jwtSecret != "" {
				authHeader := r.Header.Get("Authorization")
				if !strings.HasPrefix(authHeader, "Bearer ") {
					http.Error(w, `{"code":"UNAUTHORIZED","message":"Authorization Bearer token required"}`, http.StatusUnauthorized)
					return
				}
				tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
				claims, err := auth.ValidateToken(jwtSecret, tokenStr)
				if err != nil {
					http.Error(w, `{"code":"UNAUTHORIZED","message":"invalid or expired token"}`, http.StatusUnauthorized)
					return
				}
				ctx := context.WithValue(r.Context(), ContextKeyUserID, claims.UserID)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
			userID := r.Header.Get("x-user-id")
			if userID == "" {
				http.Error(w, `{"code":"UNAUTHORIZED","message":"x-user-id header required"}`, http.StatusUnauthorized)
				return
			}
			ctx := context.WithValue(r.Context(), ContextKeyUserID, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RecoveryMiddleware() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					log.Printf("panic recovered: %v", rec)
					http.Error(w, `{"code":"INTERNAL_ERROR","message":"internal server error"}`,
						http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

func CORSMiddleware(allowedOrigins string) Middleware {
	origins := map[string]bool{}
	for _, o := range strings.Split(allowedOrigins, ",") {
		o = strings.TrimSpace(o)
		if o != "" {
			origins[o] = true
		}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if len(origins) == 0 || origins["*"] {
				w.Header().Set("Access-Control-Allow-Origin", "*")
			} else if origins[origin] {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Vary", "Origin")
			}
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, x-user-id")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Max-Age", "86400")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func SecurityHeadersMiddleware() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("X-Frame-Options", "DENY")
			w.Header().Set("X-XSS-Protection", "1; mode=block")
			w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
			w.Header().Set("Content-Security-Policy",
				"default-src 'self'; "+
					"script-src 'self' 'unsafe-inline'; "+
					"style-src 'self' 'unsafe-inline'; "+
					"img-src 'self' data:; "+
					"connect-src 'self'; "+
					"frame-ancestors 'none'; "+
					"form-action 'self'")
			next.ServeHTTP(w, r)
		})
	}
}
