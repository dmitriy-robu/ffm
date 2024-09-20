package repository

import (
	"context"
	"fmt"
	"go-fitness/external/db"
	"go-fitness/internal/api/types"
	"time"
)

type PortalRepository struct {
	db db.SqlInterface
}

func NewPortalRepository(
	db db.SqlInterface,
) *PortalRepository {
	return &PortalRepository{
		db: db,
	}
}

func (r *PortalRepository) Store(ctx context.Context, portal types.Portal) error {
	const op = "PortalRepository.Store"

	const query = "INSERT INTO portal_videos (name,video_id,created_at,updated_at) VALUES (?,?,?,?)"

	now := time.Now()

	_, err := r.db.GetExecer().ExecContext(ctx, query, portal.Name, portal.VideoID, now, now)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
