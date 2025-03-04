package accrual

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
)

func TestSendOrder_Success(t *testing.T) {
	accrual := 500.0
	tests := []struct {
		name           string
		responseCode   int
		responseBody   string
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.responseCode)
				if tt.responseBody != "" {
					_, err := w.Write([]byte(tt.responseBody))
					if err != nil {
						t.Errorf("Failed to write response: %v", err)
						return
					}
				}
			}))
			defer server.Close()

			client := resty.New().SetBaseURL(server.URL)

			result, err := SendOrder(client, 123)

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedResult, result)
		})
	}
}

func TestSendOrder_SimpleErrors(t *testing.T) {
	tests := []struct {
		name          string
		responseCode  int
		responseBody  string
		expectedError error
	}{
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.responseCode)
				if tt.responseBody != "" {
					_, err := w.Write([]byte(tt.responseBody))
					if err != nil {
						t.Errorf("Failed to write response: %v", err)
						return
					}
				}
			}))
			defer server.Close()

			client := resty.New().SetBaseURL(server.URL)

			result, err := SendOrder(client, 123)

			assert.ErrorIs(t, err, tt.expectedError)
			assert.Nil(t, result)
		})
	}
}

func TestSendOrder_CustomErrors(t *testing.T) {
	tests := []struct {
		name          string
		responseCode  int
		retryAfter    int
		expectedError error
	}{
		{
			name:         "Too Many Requests - 429",
			responseCode: http.StatusTooManyRequests,
			retryAfter:   30,
			expectedError: &TooManyRequestsWithRetryError{
				RetryAfter: 30,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.responseCode)
				if tt.responseCode == http.StatusTooManyRequests {
					w.Header().Set("Retry-After", strconv.Itoa(tt.retryAfter))
				}
			}))
			defer server.Close()

			client := resty.New().SetBaseURL(server.URL)

			result, err := SendOrder(client, 123)

			assert.Error(t, err)
			var tooManyErr *TooManyRequestsWithRetryError
			if errors.As(err, &tooManyErr) {
				assert.Equal(t, tt.retryAfter, tooManyErr.RetryAfter)
			}
			assert.Nil(t, result)
		})
	}
}
