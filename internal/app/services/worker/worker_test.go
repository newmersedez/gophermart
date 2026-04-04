package worker

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"gophermart/internal/domain/models"
	"gophermart/internal/infrastructure/accrual"
	"gophermart/internal/infrastructure/accrual/mocks"
	storageMocks "gophermart/internal/infrastructure/storage/mocks"

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

	order := *models.NewOrder("79927398713", uuid.New())

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

	order := *models.NewOrder("79927398713", uuid.New())

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

	order := *models.NewOrder("79927398713", uuid.New())

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

	order := *models.NewOrder("79927398713", uuid.New())

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

	order := *models.NewOrder("79927398713", uuid.New())

	mockAccrual.EXPECT().
		GetOrderAccrual(mock.Anything, order.Number).
		Return(nil, errors.New("too many requests")).
		Once()

	
	worker.processOrder(context.Background(), order)
}

func TestProcessOrder_UpdateStatusError(t *testing.T) {
	mockStorage := storageMocks.NewMockStorageInterface(t)
	mockAccrual := mocks.NewMockAccrualClient(t)
	logger := slog.Default()

	worker := NewWorker(mockStorage, mockAccrual, logger)

	order := *models.NewOrder("79927398713", uuid.New())

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

	order := *models.NewOrder("79927398713", uuid.New())

	mockStorage.EXPECT().
		GetPendingOrders(mock.Anything).
		Return([]models.Order{order}, nil).
		Once()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() 

	worker.processOrders(ctx)
}

func TestStart(t *testing.T) {
	mockStorage := storageMocks.NewMockStorageInterface(t)
	mockAccrual := mocks.NewMockAccrualClient(t)
	logger := slog.Default()

	worker := NewWorker(mockStorage, mockAccrual, logger)
	worker.interval = 100 * time.Millisecond 

	ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
	defer cancel()

	
	mockStorage.EXPECT().
		GetPendingOrders(mock.Anything).
		Return([]models.Order{}, nil).
		Maybe() 

	
	worker.Start(ctx)
}
