package accrual

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"gophermart/internal/domain/models"
)

type AccrualClient interface {
	GetOrderAccrual(ctx context.Context, orderNumber string) (*AccrualResponse, error)
	MapStatus(accrualStatus string) string
}

type Client struct {
	baseURL    string
	httpClient *http.Client
	logger     *slog.Logger
}

type AccrualResponse struct {
	Order   string   `json:"order"`
	Status  string   `json:"status"`
	Accrual *float64 `json:"accrual,omitempty"`
}

func NewClient(baseURL string, logger *slog.Logger) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		logger: logger,
	}
}

func (c *Client) GetOrderAccrual(ctx context.Context, orderNumber string) (*AccrualResponse, error) {
	url := fmt.Sprintf("%s/api/orders/%s", c.baseURL, orderNumber)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request for order %s: %w", orderNumber, err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute HTTP request for order %s: %w", orderNumber, err)
	}
	defer func() { _ = resp.Body.Close() }()

	switch resp.StatusCode {
	case http.StatusOK:
		var accrualResp AccrualResponse
		if err := json.NewDecoder(resp.Body).Decode(&accrualResp); err != nil {
			return nil, fmt.Errorf("failed to decode accrual response for order %s: %w", orderNumber, err)
		}
		return &accrualResp, nil
	case http.StatusNoContent:
		return nil, nil
	case http.StatusTooManyRequests:
		retryAfter := resp.Header.Get("Retry-After")
		if retryAfter != "" {
			seconds, _ := strconv.Atoi(retryAfter)
			return nil, fmt.Errorf("too many requests, retry after %d seconds", seconds)
		}
		return nil, errors.New("too many requests")
	case http.StatusInternalServerError:
		return nil, errors.New("accrual system internal error")
	default:
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
}

func (c *Client) MapStatus(accrualStatus string) string {
	switch accrualStatus {
	case models.AccrualStatusRegistered:
		return models.OrderStatusNew
	case models.AccrualStatusProcessing:
		return models.OrderStatusProcessing
	case models.AccrualStatusInvalid:
		return models.OrderStatusInvalid
	case models.AccrualStatusProcessed:
		return models.OrderStatusProcessed
	default:
		return models.OrderStatusNew
	}
}
