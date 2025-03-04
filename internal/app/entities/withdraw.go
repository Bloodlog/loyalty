package entities

import (
	"time"
)

type Withdraw struct {
	CreatedAt time.Time
	Withdraw  float64
	UserID    int64
	ID        int
	OrderID   int
}
