package entities

import (
	"database/sql"
	"time"
)

type Order struct {
	ID          int
	UserID      int64
	OrderId     int
	StatusId    int16
	Accrual     sql.NullFloat64
	NextAttempt sql.NullTime
	Attempts    int
	CreatedAt   time.Time
	UpdatedAt   time.Time
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

func GetStatusIdByName(statusName string) (int, bool) {
	id, ok := StatusIds[statusName]
	return id, ok
}
