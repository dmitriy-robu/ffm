package types

import "time"

type Goal struct {
	ID        int
	Name      string
	VideoID   *int64
	CreatedAt time.Time
	UpdatedAt time.Time
}
