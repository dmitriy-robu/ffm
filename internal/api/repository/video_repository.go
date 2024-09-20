package repository

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"go-fitness/external/db"
	"go-fitness/internal/api/enum"
	"go-fitness/internal/api/types"
	"time"
)

type VideoRepository struct {
	db db.SqlInterface
}

func NewVideoRepository(
	db db.SqlInterface,
) *VideoRepository {
	return &VideoRepository{
		db: db,
	}
}

func (r *VideoRepository) CheckIfVideoExistByHashName(ctx context.Context, hashName string) bool {
	const op = "VideoRepository.CheckIfVideoExistByHashName"

	const query = "SELECT COUNT(*) FROM videos WHERE hash_name = ?"

	var count int

	if err := r.db.GetExecer().QueryRowContext(ctx, query, hashName).Scan(&count); err != nil {
		return false
	}

	return count > 0
}

func (r *VideoRepository) UpdatePoster(ctx context.Context, id int64, poster string) error {
	const op = "VideoRepository.UpdatePoster"

	const query = "UPDATE videos SET poster = ? WHERE id = ?"

	_, err := r.db.GetExecer().ExecContext(ctx, query, poster, id)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (r *VideoRepository) Create(ctx context.Context, video types.Video) (int64, error) {
	const op = "VideoRepository.Create"

	now := time.Now()

	video.UUID = uuid.New().String()
	video.CreatedAt = now
	video.UpdatedAt = now

	const query = `
		INSERT INTO videos (uuid,hash_name,status,duration,poster,created_at,updated_at) VALUES (?,?,?,?,?,?,?)
	`

	inId, err := r.db.GetExecer().ExecContext(ctx, query,
		video.UUID,
		video.HashName,
		video.Status,
		video.Duration,
		video.Poster,
		video.CreatedAt,
		video.UpdatedAt,
	)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	id, err := inId.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

func (r *VideoRepository) UpdateStatus(ctx context.Context, id int64, status enum.VideoStatus) error {
	const op = "VideoRepository.UpdateStatus"

	const query = "UPDATE videos SET status = ? WHERE id = ?"

	_, err := r.db.GetExecer().ExecContext(ctx, query, status, id)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (r *VideoRepository) GetByUUID(ctx context.Context, uuid string) (*types.Video, error) {
	const op = "VideoRepository.GetByUUID"

	const query = `
		SELECT id,uuid,hash_name,status,poster,created_at,updated_at FROM videos WHERE uuid = ? AND status = ?
	`

	var video types.Video

	if err := r.db.GetExecer().QueryRowContext(ctx, query, uuid, enum.VideoStatusProcessed).Scan(
		&video.ID,
		&video.UUID,
		&video.HashName,
		&video.Status,
		&video.Poster,
		&video.CreatedAt,
		&video.UpdatedAt,
	); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &video, nil
}

func (r *VideoRepository) GetListWhereStatusProcessedAndPosterIsNull(ctx context.Context) ([]types.Video, error) {
	const op = "VideoRepository.GetListWhereStatusProcessedAndPosterIsNull"

	const query = "SELECT id,uuid,hash_name,status,duration,created_at,updated_at FROM videos WHERE status = ? AND poster IS NULL"

	rows, err := r.db.GetExecer().QueryContext(ctx, query, enum.VideoStatusProcessed)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var videos []types.Video
	for rows.Next() {
		var video types.Video

		if err = rows.Scan(
			&video.ID,
			&video.UUID,
			&video.HashName,
			&video.Status,
			&video.Duration,
			&video.CreatedAt,
			&video.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		videos = append(videos, video)

	}
	return videos, nil
}

func (r *VideoRepository) GetList(ctx context.Context, filters map[string]interface{}) ([]types.Video, error) {
	const op = "VideoRepository.GetList"
	var query = "SELECT id,uuid,hash_name,status,duration,created_at,updated_at FROM videos"

	if len(filters) > 0 {
		for field, value := range filters {
			query += fmt.Sprintf(" AND %s = '%d'", field, value)
		}
	}

	rows, err := r.db.GetExecer().QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var videos []types.Video
	for rows.Next() {
		var video types.Video

		if err = rows.Scan(
			&video.ID,
			&video.UUID,
			&video.HashName,
			&video.Status,
			&video.Duration,
			&video.CreatedAt,
			&video.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		videos = append(videos, video)

	}
	return videos, nil
}

func (r *VideoRepository) Delete(ctx context.Context, id int64) error {
	const op = "VideoRepository.Delete"

	const query = "DELETE FROM videos WHERE id = ?"

	_, err := r.db.GetExecer().ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
