package entities

import (
	"time"
)

type Job struct {
	PoolAt    *time.Time
	CreatedAt time.Time
	ID        int64
	OrderID   int64
}
