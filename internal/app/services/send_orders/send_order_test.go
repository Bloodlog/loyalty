package send_orders

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
)

func TestSendOrder(t *testing.T) {
	var accrual float64
	accrual = 500

	tests := []struct {
		name           string
		responseCode   int
		responseBody   string
		retryAfter     int
		expectedError  error
		expectedResult *OrderResponse
	}{
		{
			name:         "Success - 200 OK",
			responseCode: http.StatusOK,
			responseBody: `{"order": "123", "status": "PROCESSED", "accrual": 500}`,
			expectedResult: &OrderResponse{
				Order:   "123",
				Status:  "PROCESSED",
				Accrual: &accrual,
			},
		},
		{
			name:          "No Content - 204",
			responseCode:  http.StatusNoContent,
			expectedError: ErrNoContent,
		},
		{
			name:          "Internal Server Error - 500",
			responseCode:  http.StatusInternalServerError,
			expectedError: ErrServerError,
		},
		{
			name:           "Too Many Requests - 429",
			responseCode:   http.StatusTooManyRequests,
			retryAfter:     0,
			expectedError:  ErrTooManyRequests,
			expectedResult: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.responseCode)
				if tt.responseBody != "" {
					w.Write([]byte(tt.responseBody))
				}
				if tt.responseCode == http.StatusTooManyRequests {
					retry := strconv.Itoa(tt.retryAfter)
					w.Header().Set("Retry-After", retry)
				}
			}))
			defer server.Close()

			client := resty.New().SetBaseURL(server.URL)

			result, retryAfter, err := SendOrder(client, 123)
			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
				assert.Nil(t, result)
				if errors.Is(tt.expectedError, ErrTooManyRequests) {
					assert.Equal(t, tt.retryAfter, retryAfter)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult, result)
			}
		})
	}
}
