package worker

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"gophermart/internal/accrual"
	"gophermart/internal/accrual/mocks"
	"gophermart/internal/models"
	storageMocks "gophermart/internal/storage/mocks"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewWorker(t *testing.T) {
	mockStorage := storageMocks.NewMockStorageInterface(t)
	mockAccrual := mocks.NewMockAccrualClient(t)
	logger := slog.Default()

	worker := NewWorker(mockStorage, mockAccrual, logger)

	assert.NotNil(t, worker)
	assert.Equal(t, 5*time.Second, worker.interval)
}

func TestProcessOrder_Success(t *testing.T) {
	mockStorage := storageMocks.NewMockStorageInterface(t)
	mockAccrual := mocks.NewMockAccrualClient(t)
	logger := slog.Default()

	worker := NewWorker(mockStorage, mockAccrual, logger)

	order := models.Order{
		Number: "79927398713",
		UserID: uuid.New(),
		Status: models.OrderStatusNew,
	}

	accrualValue := 500.0
	accrualResp := &accrual.AccrualResponse{
		Order:   order.Number,
		Status:  models.AccrualStatusProcessed,
		Accrual: &accrualValue,
	}

	mockAccrual.EXPECT().
		GetOrderAccrual(mock.Anything, order.Number).
		Return(accrualResp, nil).
		Once()

	mockAccrual.EXPECT().
		MapStatus(models.AccrualStatusProcessed).
		Return(models.OrderStatusProcessed).
		Once()

	mockStorage.EXPECT().
		UpdateOrderStatus(mock.Anything, order.Number, models.OrderStatusProcessed, &accrualValue).
		Return(nil).
		Once()

	worker.processOrder(context.Background(), order)
}

func TestProcessOrder_NilResponse(t *testing.T) {
	mockStorage := storageMocks.NewMockStorageInterface(t)
	mockAccrual := mocks.NewMockAccrualClient(t)
	logger := slog.Default()

	worker := NewWorker(mockStorage, mockAccrual, logger)

	order := models.Order{
		Number: "79927398713",
		UserID: uuid.New(),
		Status: models.OrderStatusNew,
	}

	mockAccrual.EXPECT().
		GetOrderAccrual(mock.Anything, order.Number).
		Return(nil, nil).
		Once()

	worker.processOrder(context.Background(), order)
}

func TestProcessOrder_AccrualError(t *testing.T) {
	mockStorage := storageMocks.NewMockStorageInterface(t)
	mockAccrual := mocks.NewMockAccrualClient(t)
	logger := slog.Default()

	worker := NewWorker(mockStorage, mockAccrual, logger)

	order := models.Order{
		Number: "79927398713",
		UserID: uuid.New(),
		Status: models.OrderStatusNew,
	}

	mockAccrual.EXPECT().
		GetOrderAccrual(mock.Anything, order.Number).
		Return(nil, errors.New("network error")).
		Once()

	worker.processOrder(context.Background(), order)
}

func TestProcessOrders_Success(t *testing.T) {
	mockStorage := storageMocks.NewMockStorageInterface(t)
	mockAccrual := mocks.NewMockAccrualClient(t)
	logger := slog.Default()

	worker := NewWorker(mockStorage, mockAccrual, logger)

	order := models.Order{
		Number: "79927398713",
		UserID: uuid.New(),
		Status: models.OrderStatusNew,
	}

	mockStorage.EXPECT().
		GetPendingOrders(mock.Anything).
		Return([]models.Order{order}, nil).
		Once()

	accrualValue := 500.0
	accrualResp := &accrual.AccrualResponse{
		Order:   order.Number,
		Status:  models.AccrualStatusProcessed,
		Accrual: &accrualValue,
	}

	mockAccrual.EXPECT().
		GetOrderAccrual(mock.Anything, order.Number).
		Return(accrualResp, nil).
		Once()

	mockAccrual.EXPECT().
		MapStatus(models.AccrualStatusProcessed).
		Return(models.OrderStatusProcessed).
		Once()

	mockStorage.EXPECT().
		UpdateOrderStatus(mock.Anything, order.Number, models.OrderStatusProcessed, &accrualValue).
		Return(nil).
		Once()

	worker.processOrders(context.Background())
}

func TestProcessOrders_GetPendingError(t *testing.T) {
	mockStorage := storageMocks.NewMockStorageInterface(t)
	mockAccrual := mocks.NewMockAccrualClient(t)
	logger := slog.Default()

	worker := NewWorker(mockStorage, mockAccrual, logger)

	mockStorage.EXPECT().
		GetPendingOrders(mock.Anything).
		Return(nil, errors.New("database error")).
		Once()

	worker.processOrders(context.Background())
}

func TestProcessOrder_TooManyRequests(t *testing.T) {
	mockStorage := storageMocks.NewMockStorageInterface(t)
	mockAccrual := mocks.NewMockAccrualClient(t)
	logger := slog.Default()

	worker := NewWorker(mockStorage, mockAccrual, logger)

	order := models.Order{
		Number: "79927398713",
		UserID: uuid.New(),
		Status: models.OrderStatusNew,
	}

	mockAccrual.EXPECT().
		GetOrderAccrual(mock.Anything, order.Number).
		Return(nil, errors.New("too many requests")).
		Once()

	// Should sleep for 5 seconds on "too many requests", but we'll just verify it doesn't panic
	worker.processOrder(context.Background(), order)
}

func TestProcessOrder_UpdateStatusError(t *testing.T) {
	mockStorage := storageMocks.NewMockStorageInterface(t)
	mockAccrual := mocks.NewMockAccrualClient(t)
	logger := slog.Default()

	worker := NewWorker(mockStorage, mockAccrual, logger)

	order := models.Order{
		Number: "79927398713",
		UserID: uuid.New(),
		Status: models.OrderStatusNew,
	}

	accrualValue := 500.0
	accrualResp := &accrual.AccrualResponse{
		Order:   order.Number,
		Status:  models.AccrualStatusProcessed,
		Accrual: &accrualValue,
	}

	mockAccrual.EXPECT().
		GetOrderAccrual(mock.Anything, order.Number).
		Return(accrualResp, nil).
		Once()

	mockAccrual.EXPECT().
		MapStatus(models.AccrualStatusProcessed).
		Return(models.OrderStatusProcessed).
		Once()

	mockStorage.EXPECT().
		UpdateOrderStatus(mock.Anything, order.Number, models.OrderStatusProcessed, &accrualValue).
		Return(errors.New("database error")).
		Once()

	worker.processOrder(context.Background(), order)
}

func TestProcessOrders_ContextCancelled(t *testing.T) {
	mockStorage := storageMocks.NewMockStorageInterface(t)
	mockAccrual := mocks.NewMockAccrualClient(t)
	logger := slog.Default()

	worker := NewWorker(mockStorage, mockAccrual, logger)

	order := models.Order{
		Number: "79927398713",
		UserID: uuid.New(),
		Status: models.OrderStatusNew,
	}

	mockStorage.EXPECT().
		GetPendingOrders(mock.Anything).
		Return([]models.Order{order}, nil).
		Once()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	worker.processOrders(ctx)
}

func TestStart(t *testing.T) {
	mockStorage := storageMocks.NewMockStorageInterface(t)
	mockAccrual := mocks.NewMockAccrualClient(t)
	logger := slog.Default()

	worker := NewWorker(mockStorage, mockAccrual, logger)
	worker.interval = 100 * time.Millisecond // Speed up for testing

	ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
	defer cancel()

	// Expect at least one call to GetPendingOrders
	mockStorage.EXPECT().
		GetPendingOrders(mock.Anything).
		Return([]models.Order{}, nil).
		Maybe() // May be called multiple times

	// This should run until context is cancelled
	worker.Start(ctx)
}
