package repository

import (
	"context"
	"fmt"
	"go-fitness/external/db"
	"go-fitness/internal/api/types"
	"time"
)

type TabRepository struct {
	db db.SqlInterface
}

func NewTabRepository(
	db db.SqlInterface,
) *TabRepository {
	return &TabRepository{
		db: db,
	}
}

func (r *TabRepository) Store(ctx context.Context, tab types.Tab) error {
	const op = "TabRepository.Store"

	const query = "INSERT INTO info_tabs (name, description, video_id, created_at, updated_at) VALUES (?, ?, ?, ?, ?)"

	now := time.Now()

	_, err := r.db.GetExecer().ExecContext(ctx, query, tab.Name, tab.Description, tab.VideoID, now, now)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (r *TabRepository) CheckIfNameExists(ctx context.Context, name string) bool {
	const op = "TabRepository.CheckIfNameExists"

	const query = "SELECT COUNT(*) FROM info_tabs WHERE name = ?"

	var count int
	err := r.db.GetExecer().QueryRowContext(ctx, query, name).Scan(&count)
	if err != nil {
		return false
	}

	return count > 0
}

func (r *TabRepository) GetByID(ctx context.Context, id int) (types.Tab, error) {
	const op = "TabRepository.GetByID"

	const query = "SELECT id, name, description, video_id FROM info_tabs WHERE id = ?"

	row := r.db.GetExecer().QueryRowContext(ctx, query, id)

	tab := types.Tab{}

	if err := row.Scan(&tab.ID, &tab.Name, &tab.Description, &tab.VideoID); err != nil {
		return tab, fmt.Errorf("%s: %w", op, err)
	}

	return tab, nil
}

func (r *TabRepository) Update(ctx context.Context, tab types.Tab) error {
	const op = "TabRepository.Update"

	const query = "UPDATE info_tabs SET name = ?, description = ?, video_id = ?, updated_at = ? WHERE id = ?"

	now := time.Now()

	_, err := r.db.GetExecer().ExecContext(ctx, query, tab.Name, tab.Description, tab.VideoID, now, tab.ID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
