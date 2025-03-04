package accrual

import (
	"time"

	"github.com/go-resty/resty/v2"
)

func NewClient(serverAddr string, timeout time.Duration) *resty.Client {
	return resty.New().
		SetBaseURL(serverAddr).
		SetHeader("Content-Type", "application/json").
		SetTimeout(timeout)
}
