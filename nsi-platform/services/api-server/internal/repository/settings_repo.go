package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/trigold786/94-AI-Insurance-Design/shared/models"
)

type settingsRepository struct {
	db *sql.DB
}

func NewSettingsRepository(db *sql.DB) (*settingsRepository, error) {
	if db == nil {
		return nil, fmt.Errorf("db cannot be nil")
	}
	return &settingsRepository{db: db}, nil
}

func (r *settingsRepository) GetSettings(ctx context.Context, userID string) (*models.SettingsData, error) {
	var s models.SettingsData
	err := r.db.QueryRowContext(ctx,
		`SELECT font_scale, default_tab, notifications_on FROM user_settings WHERE user_id = $1`, userID).
		Scan(&s.FontScale, &s.DefaultTab, &s.NotificationsOn)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &s, err
}

func (r *settingsRepository) SaveSettings(ctx context.Context, userID string, s *models.SettingsData) error {
	r.db.ExecContext(ctx, `INSERT INTO users (user_id, tenant_id) VALUES ($1, 'default') ON CONFLICT DO NOTHING`, userID)
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO user_settings (user_id, font_scale, default_tab, notifications_on) VALUES ($1, $2, $3, $4)
		 ON CONFLICT (user_id) DO UPDATE SET font_scale=EXCLUDED.font_scale, default_tab=EXCLUDED.default_tab, notifications_on=EXCLUDED.notifications_on, updated_at=NOW()`,
		userID, s.FontScale, s.DefaultTab, s.NotificationsOn)
	return err
}

func (r *settingsRepository) DeleteUserData(ctx context.Context, userID string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	tables := []string{"simulator_scenarios", "orders", "payment_records", "alerts", "plan_snapshots", "feedback", "user_settings", "user_profiles", "users"}
	for _, t := range tables {
		tx.ExecContext(ctx, fmt.Sprintf("DELETE FROM %s WHERE user_id = $1", t), userID)
	}
	return tx.Commit()
}
