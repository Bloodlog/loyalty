package entities

import (
	"database/sql"
	"time"
)

type Order struct {
	CreatedAt   time.Time
	UpdatedAt   time.Time
	NextAttempt sql.NullTime
	Accrual     sql.NullFloat64
	UserID      int64
	ID          int
	OrderID     int
	Attempts    int
	StatusID    int16
}

const (
	StatusNew        = 1 // NEW — заказ загружен, но не обработан
	StatusProcessing = 2 // PROCESSING — идет расчет вознаграждения
	StatusInvalid    = 3 // INVALID — расчет невозможен
	StatusProcessed  = 4 // PROCESSED — расчет завершен
)

var StatusIds = map[string]int{
	"NEW":        StatusNew,
	"PROCESSING": StatusProcessing,
	"INVALID":    StatusInvalid,
	"PROCESSED":  StatusProcessed,
}

var StatusNames = map[int]string{
	StatusNew:        "NEW",
	StatusProcessing: "PROCESSING",
	StatusInvalid:    "INVALID",
	StatusProcessed:  "PROCESSED",
}

func GetStatusName(statusID int) string {
	if name, ok := StatusNames[statusID]; ok {
		return name
	}
	return "UNKNOWN"
}

func GetStatusIDByName(statusName string) (int, bool) {
	id, ok := StatusIds[statusName]
	return id, ok
}
