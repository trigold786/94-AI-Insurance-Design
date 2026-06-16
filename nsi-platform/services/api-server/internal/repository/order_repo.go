package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/trigold786/94-AI-Insurance-Design/shared/models"
)

type OrderRepository interface {
	CreateOrder(ctx context.Context, orderID, userID, planID string, amount float64) error
	GetOrder(ctx context.Context, orderID string) (*models.OrderData, error)
	GetOrderByUserPlan(ctx context.Context, userID, planID string) (*models.OrderData, error)
	UpdateOrderPaid(ctx context.Context, orderID, paymentMethod string) error
	ListUserOrders(ctx context.Context, userID string) ([]models.OrderData, error)
}

type orderRepository struct {
	db *sql.DB
}

func NewOrderRepository(db *sql.DB) (OrderRepository, error) {
	if db == nil {
		return nil, fmt.Errorf("db cannot be nil")
	}
	return &orderRepository{db: db}, nil
}

func (r *orderRepository) CreateOrder(ctx context.Context, orderID, userID, planID string, amount float64) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO orders (order_id, user_id, plan_id, amount, status) VALUES ($1, $2, $3, $4, 'pending')
		 ON CONFLICT (user_id, plan_id) DO NOTHING`,
		orderID, userID, planID, amount)
	return err
}

func (r *orderRepository) GetOrder(ctx context.Context, orderID string) (*models.OrderData, error) {
	var o models.OrderData
	err := r.db.QueryRowContext(ctx,
		`SELECT order_id, user_id, plan_id, amount, status, COALESCE(payment_method,''), paid_at::text, created_at::text
		 FROM orders WHERE order_id = $1`, orderID).
		Scan(&o.OrderID, &o.UserID, &o.PlanID, &o.Amount, &o.Status, &o.PaymentMethod, &o.PaidAt, &o.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("order not found")
	}
	return &o, err
}

func (r *orderRepository) GetOrderByUserPlan(ctx context.Context, userID, planID string) (*models.OrderData, error) {
	var o models.OrderData
	err := r.db.QueryRowContext(ctx,
		`SELECT order_id, user_id, plan_id, amount, status, COALESCE(payment_method,''), paid_at::text, created_at::text
		 FROM orders WHERE user_id = $1 AND plan_id = $2`, userID, planID).
		Scan(&o.OrderID, &o.UserID, &o.PlanID, &o.Amount, &o.Status, &o.PaymentMethod, &o.PaidAt, &o.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("order not found")
	}
	return &o, err
}

func (r *orderRepository) UpdateOrderPaid(ctx context.Context, orderID, paymentMethod string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE orders SET status = 'paid', payment_method = $1, paid_at = NOW() WHERE order_id = $2`,
		paymentMethod, orderID)
	return err
}

func (r *orderRepository) ListUserOrders(ctx context.Context, userID string) ([]models.OrderData, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT order_id, user_id, plan_id, amount, status, COALESCE(payment_method,''), paid_at::text, created_at::text
		 FROM orders WHERE user_id = $1 ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var orders []models.OrderData
	for rows.Next() {
		var o models.OrderData
		if err := rows.Scan(&o.OrderID, &o.UserID, &o.PlanID, &o.Amount, &o.Status, &o.PaymentMethod, &o.PaidAt, &o.CreatedAt); err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}
	return orders, rows.Err()
}
