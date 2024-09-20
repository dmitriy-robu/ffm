package repository

import (
	"context"
	"fmt"
	"go-fitness/external/db"
	"go-fitness/internal/api/types"
	"time"
)

type WorkoutRepository struct {
	db db.SqlInterface
}

func NewWorkoutRepository(
	db db.SqlInterface,
) *WorkoutRepository {
	return &WorkoutRepository{
		db: db,
	}
}

func (r *WorkoutRepository) Update(ctx context.Context, id int64, workout types.Workout) error {
	const op = "WorkoutRepository.UpdateWorkout"

	const query = "UPDATE workouts SET name = ?, description = ?, video_id = ?, updated_at = ? WHERE id = ?"

	_, err := r.db.GetExecer().ExecContext(ctx, query, workout.Name, workout.Description, workout.VideoID, time.Now(), id)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (r *WorkoutRepository) GetWorkoutByName(ctx context.Context, name string) (types.Workout, error) {
	const op = "VideoRepository.GetWorkoutByName"

	const query = "SELECT id,name FROM workouts WHERE name = ?"

	var workout types.Workout

	if err := r.db.GetExecer().QueryRowContext(ctx, query, name).Scan(
		&workout.ID,
		&workout.Name,
	); err != nil {
		return workout, fmt.Errorf("%s: %w", op, err)
	}

	return workout, nil
}

func (r *WorkoutRepository) Create(ctx context.Context, workout types.Workout) (int64, error) {
	const op = "WorkoutRepository.CreateWorkout"

	const query = "INSERT INTO workouts (name,description,video_id,created_at,updated_at) VALUES (?,?,?,?,?)"

	now := time.Now()

	res, err := r.db.GetExecer().ExecContext(ctx, query, workout.Name, workout.Description, workout.VideoID, now, now)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

func (r *WorkoutRepository) AddWorkoutToProgramMonth(ctx context.Context, workoutID, programMonthID int64) error {
	const op = "WorkoutRepository.AddWorkoutToProgramMonth"

	const query = "INSERT INTO program_month_has_workouts (workout_id,program_month_id) VALUES (?,?)"

	_, err := r.db.GetExecer().ExecContext(ctx, query, workoutID, programMonthID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (r *WorkoutRepository) CheckIfNameExists(ctx context.Context, name string) bool {
	const op = "WorkoutRepository.CheckIfNameExists"

	const query = "SELECT COUNT(id) FROM workouts WHERE name = ?"

	var count int

	if err := r.db.GetExecer().QueryRowContext(ctx, query, name).Scan(&count); err != nil {
		return false
	}

	return count > 0
}
