package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/trigold786/94-AI-Insurance-Design/shared/middleware"
	"github.com/trigold786/94-AI-Insurance-Design/shared/models"
)

type OrderRepository interface {
	CreateOrder(ctx context.Context, orderID, userID, planID string, amount float64) error
	GetOrder(ctx context.Context, orderID string) (*models.OrderData, error)
	GetOrderByUserPlan(ctx context.Context, userID, planID string) (*models.OrderData, error)
	UpdateOrderPaid(ctx context.Context, orderID, paymentMethod string) error
	ListUserOrders(ctx context.Context, userID string) ([]models.OrderData, error)
}

func CreateOrderHandler(orderRepo OrderRepository, planRepo PlanRepository) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, _ := r.Context().Value(middleware.ContextKeyUserID).(string)

		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
		var req struct {
			PlanID string `json:"plan_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondJSON(w, 400, map[string]interface{}{"code": "VALIDATION_ERROR", "message": "invalid JSON"})
			return
		}
		if req.PlanID == "" {
			respondJSON(w, 400, map[string]interface{}{"code": "VALIDATION_ERROR", "message": "plan_id required"})
			return
		}

		plan, err := planRepo.GetByID(r.Context(), req.PlanID)
		if err != nil {
			respondJSON(w, 404, map[string]interface{}{"code": "NOT_FOUND", "message": "plan not found"})
			return
		}
		if plan.UserID != userID {
			respondJSON(w, 403, map[string]interface{}{"code": "FORBIDDEN", "message": "plan does not belong to user"})
			return
		}

		if existing, err := orderRepo.GetOrderByUserPlan(r.Context(), userID, req.PlanID); err == nil && existing != nil {
			if existing.Status == "paid" {
				respondJSON(w, 200, map[string]interface{}{"code": 0, "data": existing, "message": "already paid"})
				return
			}
			respondJSON(w, 200, map[string]interface{}{"code": 0, "data": existing, "message": "order already exists"})
			return
		}

		amount := 19.90
		orderID := fmt.Sprintf("ord-%d", time.Now().UnixNano())
		orderRepo.CreateOrder(r.Context(), orderID, userID, req.PlanID, amount)

		order, _ := orderRepo.GetOrderByUserPlan(r.Context(), userID, req.PlanID)
		respondJSON(w, 200, map[string]interface{}{"code": 0, "data": order})
	})
}

func PayOrderHandler(orderRepo OrderRepository) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, _ := r.Context().Value(middleware.ContextKeyUserID).(string)

		orderID := strings.TrimPrefix(r.URL.Path, "/v1/orders/")
		orderID = strings.TrimSuffix(orderID, "/pay")
		if orderID == "" {
			respondJSON(w, 400, map[string]interface{}{"code": "VALIDATION_ERROR", "message": "order_id required"})
			return
		}

		order, err := orderRepo.GetOrder(r.Context(), orderID)
		if err != nil || order == nil {
			respondJSON(w, 404, map[string]interface{}{"code": "NOT_FOUND", "message": "order not found"})
			return
		}
		if order.UserID != userID {
			respondJSON(w, 403, map[string]interface{}{"code": "FORBIDDEN", "message": "order does not belong to user"})
			return
		}
		if order.Status == "paid" {
			respondJSON(w, 200, map[string]interface{}{"code": 0, "data": order, "message": "already paid"})
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
		var req struct {
			PaymentMethod string `json:"payment_method"`
		}
		json.NewDecoder(r.Body).Decode(&req)
		if req.PaymentMethod == "" {
			req.PaymentMethod = "wechat"
		}

		if err := orderRepo.UpdateOrderPaid(r.Context(), orderID, req.PaymentMethod); err != nil {
			respondError(w, err)
			return
		}

		order, _ = orderRepo.GetOrder(r.Context(), orderID)
		respondJSON(w, 200, map[string]interface{}{"code": 0, "data": order, "message": "payment successful"})
	})
}

func CheckUnlockHandler(orderRepo OrderRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, _ := r.Context().Value(middleware.ContextKeyUserID).(string)
		planID := r.URL.Query().Get("plan_id")
		if planID == "" {
			respondJSON(w, 400, map[string]interface{}{"code": "VALIDATION_ERROR", "message": "plan_id required"})
			return
		}
		unlocked := false
		if order, err := orderRepo.GetOrderByUserPlan(r.Context(), userID, planID); err == nil && order != nil {
			unlocked = order.Status == "paid"
		}
		respondJSON(w, 200, map[string]interface{}{"code": 0, "data": map[string]bool{"unlocked": unlocked}})
	}
}
