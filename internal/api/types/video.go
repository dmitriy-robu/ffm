package types

import (
	"go-fitness/internal/api/enum"
	"time"
)

type Video struct {
	ID        int64
	UUID      string
	HashName  string
	Status    enum.VideoStatus
	Duration  float64
	Poster    *string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type VideoPosition struct {
	ID        int64
	UserID    int64
	VideoID   int64
	Position  float64
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Workout struct {
	ID          int64
	Name        string
	Description string
	VideoID     int64
	DeletedAt   *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
