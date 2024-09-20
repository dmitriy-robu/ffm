package repository

import (
	"context"
	"fmt"
	"go-fitness/external/db"
	"go-fitness/internal/api/types"
	"time"
)

type GoalRepository struct {
	db db.SqlInterface
}

func NewGoalRepository(
	db db.SqlInterface,
) *GoalRepository {
	return &GoalRepository{
		db: db,
	}
}

func (r *GoalRepository) Store(ctx context.Context, goal types.Goal) error {
	const op = "GoalRepository.Store"

	const query = "INSERT INTO goals (name, video_id, created_at, updated_at) VALUES (?, ?, ?, ?)"

	now := time.Now()

	_, err := r.db.GetExecer().ExecContext(ctx, query, goal.Name, goal.VideoID, now, now)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (r *GoalRepository) CheckIfNameExists(ctx context.Context, name string) bool {
	const op = "GoalRepository.CheckIfNameExists"

	const query = "SELECT COUNT(*) FROM goals WHERE name = ?"

	var count int
	err := r.db.GetExecer().QueryRowContext(ctx, query, name).Scan(&count)
	if err != nil {
		return false
	}

	return count > 0
}

func (r *GoalRepository) GetByID(ctx context.Context, id int) (types.Goal, error) {
	const op = "GoalRepository.GetByID"

	const query = "SELECT id, name, video_id FROM goals WHERE id = ?"

	row := r.db.GetExecer().QueryRowContext(ctx, query, id)

	goal := types.Goal{}

	if err := row.Scan(&goal.ID, &goal.Name, &goal.VideoID); err != nil {
		return goal, fmt.Errorf("%s: %w", op, err)
	}

	return goal, nil
}

func (r *GoalRepository) Update(ctx context.Context, goal types.Goal) error {
	const op = "GoalRepository.Update"

	const query = "UPDATE goals SET name = ?, video_id = ?, updated_at = ? WHERE id = ?"

	now := time.Now()

	_, err := r.db.GetExecer().ExecContext(ctx, query, goal.Name, goal.VideoID, now, goal.ID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
