package accrual

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"gophermart/internal/domain/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	client := NewClient("http://localhost:8081", logger)

	require.NotNil(t, client)
	assert.Equal(t, "http://localhost:8081", client.baseURL)
	assert.NotNil(t, client.httpClient)
	assert.NotNil(t, client.logger)
}

func TestGetOrderAccrual_Success(t *testing.T) {
	accrual := 500.0
	response := AccrualResponse{
		Order:   "12345678903",
		Status:  models.AccrualStatusProcessed,
		Accrual: &accrual,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/orders/12345678903", r.URL.Path)
		assert.Equal(t, http.MethodGet, r.Method)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	client := NewClient(server.URL, logger)

	result, err := client.GetOrderAccrual(context.Background(), "12345678903")

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "12345678903", result.Order)
	assert.Equal(t, models.AccrualStatusProcessed, result.Status)
	assert.Equal(t, 500.0, *result.Accrual)
}

func TestGetOrderAccrual_NoContent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	client := NewClient(server.URL, logger)

	result, err := client.GetOrderAccrual(context.Background(), "12345678903")

	require.NoError(t, err)
	assert.Nil(t, result)
}

func TestGetOrderAccrual_TooManyRequests(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "60")
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	client := NewClient(server.URL, logger)

	result, err := client.GetOrderAccrual(context.Background(), "12345678903")

	require.Error(t, err)
	assert.Nil(t, result)
	assert.ErrorContains(t, err, "too many requests")
	assert.ErrorContains(t, err, "60 seconds")
}

func TestGetOrderAccrual_TooManyRequestsNoRetryAfter(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	client := NewClient(server.URL, logger)

	result, err := client.GetOrderAccrual(context.Background(), "12345678903")

	require.Error(t, err)
	assert.Nil(t, result)
	assert.EqualError(t, err, "too many requests")
}

func TestGetOrderAccrual_InternalServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	client := NewClient(server.URL, logger)

	result, err := client.GetOrderAccrual(context.Background(), "12345678903")

	require.Error(t, err)
	assert.Nil(t, result)
	assert.EqualError(t, err, "accrual system internal error")
}

func TestGetOrderAccrual_UnexpectedStatusCode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	client := NewClient(server.URL, logger)

	result, err := client.GetOrderAccrual(context.Background(), "12345678903")

	require.Error(t, err)
	assert.Nil(t, result)
	assert.ErrorContains(t, err, "unexpected status code: 400")
}

func TestMapStatus(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	client := NewClient("http://localhost:8081", logger)

	tests := []struct {
		name           string
		accrualStatus  string
		expectedStatus string
	}{
		{
			name:           "REGISTERED maps to NEW",
			accrualStatus:  models.AccrualStatusRegistered,
			expectedStatus: models.OrderStatusNew,
		},
		{
			name:           "PROCESSING maps to PROCESSING",
			accrualStatus:  models.AccrualStatusProcessing,
			expectedStatus: models.OrderStatusProcessing,
		},
		{
			name:           "INVALID maps to INVALID",
			accrualStatus:  models.AccrualStatusInvalid,
			expectedStatus: models.OrderStatusInvalid,
		},
		{
			name:           "PROCESSED maps to PROCESSED",
			accrualStatus:  models.AccrualStatusProcessed,
			expectedStatus: models.OrderStatusProcessed,
		},
		{
			name:           "Unknown status maps to NEW",
			accrualStatus:  "UNKNOWN",
			expectedStatus: models.OrderStatusNew,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := client.MapStatus(tt.accrualStatus)
			assert.Equal(t, tt.expectedStatus, result)
		})
	}
}

func TestGetOrderAccrual_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("invalid json"))
	}))
	defer server.Close()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	client := NewClient(server.URL, logger)

	resp, err := client.GetOrderAccrual(context.Background(), "12345678903")
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestGetOrderAccrual_TooManyRequestsInvalidRetryAfter(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "not-a-number")
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	client := NewClient(server.URL, logger)

	resp, err := client.GetOrderAccrual(context.Background(), "12345678903")
	assert.Error(t, err)
	assert.ErrorContains(t, err, "retry after 0 seconds")
	assert.Nil(t, resp)
}
