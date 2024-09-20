package types

import "time"

type Tab struct {
	ID          int
	Name        string
	Description string
	VideoID     *int64
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
