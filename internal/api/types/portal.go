package types

import "time"

type Portal struct {
	ID        int
	Name      string
	VideoID   int64
	CreatedAt time.Time
	UpdatedAt time.Time
}
