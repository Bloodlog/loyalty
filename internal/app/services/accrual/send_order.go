package accrual

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-resty/resty/v2"

	"strconv"
)

type OrderResponse struct {
	Accrual *float64 `json:"accrual,omitempty"`
	Order   string   `json:"order"`
	Status  string   `json:"status"`
}

var (
	ErrTooManyRequests = errors.New("too many requests (429)")
	ErrServerError     = errors.New("internal server error (500)")
	ErrNoContent       = errors.New("no content (204)")
)

func SendOrder(client *resty.Client, orderID int64) (*OrderResponse, int, error) {
	resp, err := client.R().
		Get("/api/orders/" + strconv.Itoa(int(orderID)))

	if err != nil {
		return nil, 0, fmt.Errorf("failed to send order: %w", err)
	}
	switch resp.StatusCode() {
	case http.StatusOK:
		var orderResp OrderResponse
		if err := json.Unmarshal(resp.Body(), &orderResp); err != nil {
			return nil, 0, fmt.Errorf("failed to parse response: %w", err)
		}
		return &orderResp, 0, nil
	case http.StatusNoContent:
		return nil, 0, ErrNoContent
	case http.StatusInternalServerError:
		return nil, 0, ErrServerError
	case http.StatusTooManyRequests:
		retryAfter := 0
		if header := resp.Header().Get("Retry-After"); header != "" {
			if seconds, err := strconv.Atoi(header); err == nil {
				retryAfter = seconds
			}
		}
		return nil, retryAfter, ErrTooManyRequests
	default:
		return nil, 0, fmt.Errorf("unexpected status code: %d", resp.StatusCode())
	}
}
