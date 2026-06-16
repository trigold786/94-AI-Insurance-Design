package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/trigold786/94-AI-Insurance-Design/shared/models"
)

type SimulatorScenarioRepository interface {
	SaveScenario(ctx context.Context, userID, name string, params, result json.RawMessage) error
	ListScenarios(ctx context.Context, userID string) ([]models.SimScenario, error)
}

type simulatorScenarioRepository struct {
	db *sql.DB
}

func NewSimulatorScenarioRepository(db *sql.DB) (SimulatorScenarioRepository, error) {
	if db == nil {
		return nil, fmt.Errorf("db cannot be nil")
	}
	return &simulatorScenarioRepository{db: db}, nil
}

func (r *simulatorScenarioRepository) SaveScenario(ctx context.Context, userID, name string, params, result json.RawMessage) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	rows, err := tx.QueryContext(ctx, "SELECT id FROM simulator_scenarios WHERE user_id = $1 FOR UPDATE", userID)
	if err != nil {
		return err
	}
	var ids []int
	for rows.Next() {
		var id int
		rows.Scan(&id)
		ids = append(ids, id)
	}
	rows.Close()

	if len(ids) >= 3 {
		_, err := tx.ExecContext(ctx, "DELETE FROM simulator_scenarios WHERE id = $1", ids[0])
		if err != nil {
			return err
		}
	}
	_, err = tx.ExecContext(ctx,
		`INSERT INTO simulator_scenarios (user_id, name, params, result) VALUES ($1, $2, $3::jsonb, $4::jsonb)`,
		userID, name, string(params), string(result))
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (r *simulatorScenarioRepository) ListScenarios(ctx context.Context, userID string) ([]models.SimScenario, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, user_id, name, params, COALESCE(result,'{}'::jsonb), created_at::text
		 FROM simulator_scenarios WHERE user_id = $1 ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var scenarios []models.SimScenario
	for rows.Next() {
		var s models.SimScenario
		if err := rows.Scan(&s.ID, &s.UserID, &s.Name, &s.Params, &s.Result, &s.CreatedAt); err != nil {
			return nil, err
		}
		scenarios = append(scenarios, s)
	}
	return scenarios, rows.Err()
}
