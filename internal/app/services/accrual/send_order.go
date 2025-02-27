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
	ErrServerError = errors.New("internal server error (500)")
	ErrNoContent   = errors.New("no content (204)")
)

type TooManyRequestsWithRetryError struct {
	RetryAfter int
}

func (e *TooManyRequestsWithRetryError) Error() string {
	return fmt.Sprintf("too many requests (429), retry after %d seconds", e.RetryAfter)
}

func SendOrder(client *resty.Client, orderID int64) (*OrderResponse, error) {
	resp, err := client.R().
		Get("/api/orders/" + strconv.Itoa(int(orderID)))

	if err != nil {
		return nil, fmt.Errorf("failed to send order: %w", err)
	}
	switch resp.StatusCode() {
	case http.StatusOK:
		return handleStatusOkRequests(resp)
	case http.StatusNoContent:
		return nil, ErrNoContent
	case http.StatusInternalServerError:
		return nil, ErrServerError
	case http.StatusTooManyRequests:
		return nil, handleTooManyRequests(resp)
	default:
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode())
	}
}

func handleStatusOkRequests(resp *resty.Response) (*OrderResponse, error) {
	var orderResp OrderResponse
	if err := json.Unmarshal(resp.Body(), &orderResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &orderResp, nil
}

func handleTooManyRequests(resp *resty.Response) error {
	retryAfter := 30
	if header := resp.Header().Get("Retry-After"); header != "" {
		if seconds, err := strconv.Atoi(header); err == nil {
			retryAfter = seconds
		}
	}
	return &TooManyRequestsWithRetryError{RetryAfter: retryAfter}
}
