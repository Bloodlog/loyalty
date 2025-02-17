package entities

import (
	"time"
)

type Withdraw struct {
	ID        int
	UserID    int64
	OrderId   int
	Withdraw  float64
	CreatedAt time.Time
}
