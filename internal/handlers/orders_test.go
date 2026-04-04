package handlers

import (
	"bytes"
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"gophermart/internal/middleware"
	"gophermart/internal/models"
	"gophermart/internal/storage/mocks"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUploadOrder_Success(t *testing.T) {
	logger := slog.Default()
	mockStorage := mocks.NewMockStorageInterface(t)

	userID := uuid.New()
	orderNumber := "79927398713"

	mockStorage.EXPECT().
		GetOrderByNumber(mock.Anything, orderNumber).
		Return(nil, nil).
		Once()

	mockStorage.EXPECT().
		CreateOrder(mock.Anything, orderNumber, userID).
		Return(nil).
		Once()

	handler := NewOrderHandler(mockStorage, logger)

	request := httptest.NewRequest(http.MethodPost, "/api/user/orders", bytes.NewReader([]byte(orderNumber)))
	request.Header.Set("Content-Type", "text/plain")

	ctx := context.WithValue(request.Context(), middleware.UserIDKey, userID)
	request = request.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.UploadOrder(w, request)

	res := w.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusAccepted, res.StatusCode)
}

func TestUploadOrder_InvalidLuhn(t *testing.T) {
	logger := slog.Default()
	mockStorage := mocks.NewMockStorageInterface(t)
	handler := NewOrderHandler(mockStorage, logger)

	userID := uuid.New()
	orderNumber := "12345"

	request := httptest.NewRequest(http.MethodPost, "/api/user/orders", bytes.NewReader([]byte(orderNumber)))
	request.Header.Set("Content-Type", "text/plain")

	ctx := context.WithValue(request.Context(), middleware.UserIDKey, userID)
	request = request.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.UploadOrder(w, request)

	res := w.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusUnprocessableEntity, res.StatusCode)
}

func TestGetOrders_Success(t *testing.T) {
	logger := slog.Default()
	mockStorage := mocks.NewMockStorageInterface(t)

	userID := uuid.New()
	now := time.Now()
	accrual := 500.0

	orders := []models.Order{
		{
			Number:     "79927398713",
			UserID:     userID,
			Status:     models.OrderStatusProcessed,
			Accrual:    &accrual,
			UploadedAt: now,
		},
	}

	mockStorage.EXPECT().
		GetUserOrders(mock.Anything, userID).
		Return(orders, nil).
		Once()

	handler := NewOrderHandler(mockStorage, logger)

	request := httptest.NewRequest(http.MethodGet, "/api/user/orders", nil)

	ctx := context.WithValue(request.Context(), middleware.UserIDKey, userID)
	request = request.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.GetOrders(w, request)

	res := w.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusOK, res.StatusCode)
}

func TestGetOrders_NoOrders(t *testing.T) {
	logger := slog.Default()
	mockStorage := mocks.NewMockStorageInterface(t)

	userID := uuid.New()

	mockStorage.EXPECT().
		GetUserOrders(mock.Anything, userID).
		Return([]models.Order{}, nil).
		Once()

	handler := NewOrderHandler(mockStorage, logger)

	request := httptest.NewRequest(http.MethodGet, "/api/user/orders", nil)

	ctx := context.WithValue(request.Context(), middleware.UserIDKey, userID)
	request = request.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.GetOrders(w, request)

	res := w.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusNoContent, res.StatusCode)
}

func TestUploadOrder_EmptyBody(t *testing.T) {
	logger := slog.Default()
	mockStorage := mocks.NewMockStorageInterface(t)
	handler := NewOrderHandler(mockStorage, logger)

	userID := uuid.New()

	request := httptest.NewRequest(http.MethodPost, "/api/user/orders", bytes.NewReader([]byte("")))
	ctx := context.WithValue(request.Context(), middleware.UserIDKey, userID)
	request = request.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.UploadOrder(w, request)

	res := w.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestUploadOrder_AlreadyExistsSameUser(t *testing.T) {
	logger := slog.Default()
	mockStorage := mocks.NewMockStorageInterface(t)

	userID := uuid.New()
	orderNumber := "79927398713"

	existingOrder := &models.Order{
		Number: orderNumber,
		UserID: userID,
		Status: models.OrderStatusNew,
	}

	mockStorage.EXPECT().
		GetOrderByNumber(mock.Anything, orderNumber).
		Return(existingOrder, nil).
		Once()

	handler := NewOrderHandler(mockStorage, logger)

	request := httptest.NewRequest(http.MethodPost, "/api/user/orders", bytes.NewReader([]byte(orderNumber)))
	ctx := context.WithValue(request.Context(), middleware.UserIDKey, userID)
	request = request.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.UploadOrder(w, request)

	res := w.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusOK, res.StatusCode)
}

func TestUploadOrder_AlreadyExistsDifferentUser(t *testing.T) {
	logger := slog.Default()
	mockStorage := mocks.NewMockStorageInterface(t)

	userID := uuid.New()
	otherUserID := uuid.New()
	orderNumber := "79927398713"

	existingOrder := &models.Order{
		Number: orderNumber,
		UserID: otherUserID,
		Status: models.OrderStatusNew,
	}

	mockStorage.EXPECT().
		GetOrderByNumber(mock.Anything, orderNumber).
		Return(existingOrder, nil).
		Once()

	handler := NewOrderHandler(mockStorage, logger)

	request := httptest.NewRequest(http.MethodPost, "/api/user/orders", bytes.NewReader([]byte(orderNumber)))
	ctx := context.WithValue(request.Context(), middleware.UserIDKey, userID)
	request = request.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.UploadOrder(w, request)

	res := w.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusConflict, res.StatusCode)
}

func TestUploadOrder_GetOrderError(t *testing.T) {
	logger := slog.Default()
	mockStorage := mocks.NewMockStorageInterface(t)

	userID := uuid.New()
	orderNumber := "79927398713"

	mockStorage.EXPECT().
		GetOrderByNumber(mock.Anything, orderNumber).
		Return(nil, assert.AnError).
		Once()

	handler := NewOrderHandler(mockStorage, logger)

	request := httptest.NewRequest(http.MethodPost, "/api/user/orders", bytes.NewReader([]byte(orderNumber)))
	ctx := context.WithValue(request.Context(), middleware.UserIDKey, userID)
	request = request.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.UploadOrder(w, request)

	res := w.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusInternalServerError, res.StatusCode)
}

func TestUploadOrder_CreateOrderError(t *testing.T) {
	logger := slog.Default()
	mockStorage := mocks.NewMockStorageInterface(t)

	userID := uuid.New()
	orderNumber := "79927398713"

	mockStorage.EXPECT().
		GetOrderByNumber(mock.Anything, orderNumber).
		Return(nil, nil).
		Once()

	mockStorage.EXPECT().
		CreateOrder(mock.Anything, orderNumber, userID).
		Return(assert.AnError).
		Once()

	handler := NewOrderHandler(mockStorage, logger)

	request := httptest.NewRequest(http.MethodPost, "/api/user/orders", bytes.NewReader([]byte(orderNumber)))
	ctx := context.WithValue(request.Context(), middleware.UserIDKey, userID)
	request = request.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.UploadOrder(w, request)

	res := w.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusInternalServerError, res.StatusCode)
}

func TestGetOrders_Error(t *testing.T) {
	logger := slog.Default()
	mockStorage := mocks.NewMockStorageInterface(t)

	userID := uuid.New()

	mockStorage.EXPECT().
		GetUserOrders(mock.Anything, userID).
		Return(nil, assert.AnError).
		Once()

	handler := NewOrderHandler(mockStorage, logger)

	request := httptest.NewRequest(http.MethodGet, "/api/user/orders", nil)
	ctx := context.WithValue(request.Context(), middleware.UserIDKey, userID)
	request = request.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.GetOrders(w, request)

	res := w.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusInternalServerError, res.StatusCode)
}
