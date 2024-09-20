package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"go-fitness/external/db"
	"go-fitness/internal/api/types"
	"time"
)

type ProgramRepository struct {
	db db.SqlInterface
}

func NewProgramRepository(
	db db.SqlInterface,
) *ProgramRepository {
	return &ProgramRepository{
		db: db,
	}
}

func (r *ProgramRepository) GetProgramMonth(ctx context.Context, programID int64, month int) (types.ProgramMonth, error) {
	const op string = "repository.ProgramRepository.GetProgramMonth"

	const query = "SELECT id, program_id, month FROM program_months WHERE program_id = ? AND month = ?"

	row := r.db.GetExecer().QueryRowContext(ctx, query, programID, month)

	programMonth := types.ProgramMonth{}

	if err := row.Scan(&programMonth.ID, &programMonth.ProgramID, &programMonth.Month); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return programMonth, fmt.Errorf("%s: %w", op, err)
		}

		return programMonth, fmt.Errorf("%s: %w", op, err)
	}

	return programMonth, nil
}

func (r *ProgramRepository) CreateProgramMonth(ctx context.Context, programID int64, month int) (int64, error) {
	const op string = "repository.ProgramRepository.CreateProgramMonth"

	const query = "INSERT INTO program_months (program_id, month,created_at,updated_at) VALUES (?, ?,?,?)"

	now := time.Now()

	res, err := r.db.GetExecer().ExecContext(ctx, query, programID, month, now, now)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	id, _ := res.LastInsertId()

	return id, nil
}

func (r *ProgramRepository) GetProgramByName(ctx context.Context, name string) (types.Program, error) {
	const op string = "repository.ProgramRepository.GetProgramByData"

	const query = "SELECT id, name FROM programs WHERE name = ?"

	row := r.db.GetExecer().QueryRowContext(ctx, query, name)

	program := types.Program{}

	err := row.Scan(&program.ID, &program.Name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return program, fmt.Errorf("%s: %w", op, err)
		}

		return program, fmt.Errorf("%s: %w", op, err)
	}

	return program, nil
}

func (r *ProgramRepository) GetGoalByName(ctx context.Context, name string) (types.Goal, error) {
	const op string = "repository.ProgramRepository.GetGoalByID"

	const query = "SELECT id, name FROM goals WHERE name = ?"

	row := r.db.GetExecer().QueryRowContext(ctx, query, name)

	goal := types.Goal{}

	err := row.Scan(&goal.ID, &goal.Name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return goal, fmt.Errorf("%s: %w", op, err)
		}

		return goal, fmt.Errorf("%s: %w", op, err)
	}

	return goal, nil
}

func (r *ProgramRepository) GetLevelByName(ctx context.Context, name string) (types.Level, error) {
	const op string = "repository.ProgramRepository.GetLevelByID"

	const query = "SELECT id, name FROM levels WHERE name = ?"

	row := r.db.GetExecer().QueryRowContext(ctx, query, name)

	level := types.Level{}

	err := row.Scan(&level.ID, &level.Name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return level, fmt.Errorf("%s: %w", op, err)
		}

		return level, fmt.Errorf("%s: %w", op, err)
	}

	return level, nil
}

func (r *ProgramRepository) GetPeriodByName(ctx context.Context, name string) (types.Period, error) {
	const op string = "repository.ProgramRepository.GetPeriodByID"

	const query = "SELECT id, name FROM periods WHERE name = ?"

	row := r.db.GetExecer().QueryRowContext(ctx, query, name)

	period := types.Period{}

	err := row.Scan(&period.ID, &period.Name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return period, fmt.Errorf("%s: %w", op, err)
		}

		return period, fmt.Errorf("%s: %w", op, err)
	}

	return period, nil
}
