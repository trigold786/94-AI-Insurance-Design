package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/trigold786/94-AI-Insurance-Design/shared/errors"
	"github.com/trigold786/94-AI-Insurance-Design/shared/models"
)

type RightsRepository interface {
	GetPaymentRecords(ctx context.Context, userID string) ([]models.PaymentRecord, error)
	GetAlerts(ctx context.Context, userID string, unreadOnly bool) ([]models.Alert, error)
	MarkAlertRead(ctx context.Context, alertID, userID string) error
	CreateAlert(ctx context.Context, alert *models.Alert) error
	GetDueSoonPayments(ctx context.Context, withinDays int) ([]models.PaymentRecord, error)
	UpsertPaymentRecord(ctx context.Context, record *models.PaymentRecord) error
}

type rightsRepository struct {
	db *sql.DB
}

func NewRightsRepository(db *sql.DB) (RightsRepository, error) {
	if db == nil {
		return nil, fmt.Errorf("db cannot be nil")
	}
	return &rightsRepository{db: db}, nil
}

func (r *rightsRepository) GetPaymentRecords(ctx context.Context, userID string) ([]models.PaymentRecord, error) {
	query := `SELECT record_id, user_id, policy_type, month, amount, status, due_date, COALESCE(paid_date, '')
		FROM payment_records WHERE user_id = $1 ORDER BY month DESC, policy_type`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, errors.NewInternal("failed to query payments")
	}
	defer rows.Close()

	var records []models.PaymentRecord
	for rows.Next() {
		var rec models.PaymentRecord
		if err := rows.Scan(&rec.RecordID, &rec.UserID, &rec.PolicyType, &rec.Month, &rec.Amount, &rec.Status, &rec.DueDate, &rec.PaidDate); err != nil {
			return nil, errors.NewInternal("failed to scan payment")
		}
		records = append(records, rec)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.NewInternal("rows iteration error")
	}

	if records == nil {
		records = []models.PaymentRecord{}
	}
	return records, nil
}

func (r *rightsRepository) GetAlerts(ctx context.Context, userID string, unreadOnly bool) ([]models.Alert, error) {
	query := `SELECT alert_id, user_id, alert_type, severity, title, message, is_read, created_at::text, COALESCE(policy_id, '')
		FROM alerts WHERE user_id = $1`
	if unreadOnly {
		query += ` AND NOT is_read`
	}
	query += ` ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, errors.NewInternal("failed to query alerts")
	}
	defer rows.Close()

	var alerts []models.Alert
	for rows.Next() {
		var a models.Alert
		if err := rows.Scan(&a.AlertID, &a.UserID, &a.AlertType, &a.Severity, &a.Title, &a.Message, &a.IsRead, &a.CreatedAt, &a.PolicyID); err != nil {
			return nil, errors.NewInternal("failed to scan alert")
		}
		alerts = append(alerts, a)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.NewInternal("rows iteration error")
	}

	if alerts == nil {
		alerts = []models.Alert{}
	}
	return alerts, nil
}

func (r *rightsRepository) CreateAlert(ctx context.Context, alert *models.Alert) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO alerts (alert_id, user_id, alert_type, severity, title, message, policy_id)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 ON CONFLICT (alert_id) DO NOTHING`,
		alert.AlertID, alert.UserID, alert.AlertType, alert.Severity,
		alert.Title, alert.Message, alert.PolicyID)
	return err
}

func (r *rightsRepository) GetDueSoonPayments(ctx context.Context, withinDays int) ([]models.PaymentRecord, error) {
	query := `SELECT record_id, user_id, policy_type, month, amount, status, due_date, COALESCE(paid_date, '')
		FROM payment_records
		WHERE status = 'pending'
		AND due_date <= CURRENT_DATE + $1::integer
		AND due_date >= CURRENT_DATE
		ORDER BY due_date ASC`
	rows, err := r.db.QueryContext(ctx, query, withinDays)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var records []models.PaymentRecord
	for rows.Next() {
		var rec models.PaymentRecord
		if err := rows.Scan(&rec.RecordID, &rec.UserID, &rec.PolicyType, &rec.Month, &rec.Amount, &rec.Status, &rec.DueDate, &rec.PaidDate); err != nil {
			return nil, err
		}
		records = append(records, rec)
	}
	return records, rows.Err()
}

func (r *rightsRepository) UpsertPaymentRecord(ctx context.Context, record *models.PaymentRecord) error {
	query := `INSERT INTO payment_records (record_id, user_id, policy_type, month, amount, status, due_date, paid_date)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (record_id) DO UPDATE SET
			policy_type = EXCLUDED.policy_type,
			month = EXCLUDED.month,
			amount = EXCLUDED.amount,
			status = EXCLUDED.status,
			due_date = EXCLUDED.due_date,
			paid_date = EXCLUDED.paid_date`
	_, err := r.db.ExecContext(ctx, query,
		record.RecordID, record.UserID, record.PolicyType, record.Month,
		record.Amount, record.Status, record.DueDate, record.PaidDate)
	if err != nil {
		return errors.NewInternal("failed to upsert payment record")
	}
	return nil
}

func (r *rightsRepository) MarkAlertRead(ctx context.Context, alertID, userID string) error {
	result, err := r.db.ExecContext(ctx, `UPDATE alerts SET is_read = TRUE WHERE alert_id = $1 AND user_id = $2`, alertID, userID)
	if err != nil {
		return errors.NewInternal("failed to mark alert read")
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return errors.NewNotFound("alert", alertID)
	}
	return nil
}
