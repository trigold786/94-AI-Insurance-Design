package repository

import (
	"context"
	"database/sql"
	"fmt"
)

type FeedbackRepo struct {
	db *sql.DB
}

func NewFeedbackRepo(db *sql.DB) (*FeedbackRepo, error) {
	if db == nil {
		return nil, fmt.Errorf("db cannot be nil")
	}
	return &FeedbackRepo{db: db}, nil
}

func (r *FeedbackRepo) SaveFeedback(ctx context.Context, userID, category, content, contact string) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO feedback (user_id, category, content, contact) VALUES ($1, $2, $3, $4)`,
		userID, category, content, contact)
	return err
}
