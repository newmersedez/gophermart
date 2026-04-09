package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"gophermart/internal/app/middleware"
	"gophermart/internal/domain/models"
	"gophermart/internal/infrastructure/storage"
	"gophermart/internal/infrastructure/storage/mocks"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetBalance_Success(t *testing.T) {
	logger := slog.Default()
	mockStorage := mocks.NewMockStorageInterface(t)

	userID := uuid.New()
	balance := models.NewBalance(1000.50, 250.0)

	mockStorage.EXPECT().
		GetBalance(mock.Anything, userID).
		Return(balance, nil).
		Once()

	handler := NewBalanceHandler(mockStorage, logger)

	request := httptest.NewRequest(http.MethodGet, "/api/user/balance", nil)

	ctx := middleware.SetUserID(request.Context(), userID)
	request = request.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.GetBalance(w, request)

	res := w.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusOK, res.StatusCode)
}

func TestGetBalance_StorageError(t *testing.T) {
	logger := slog.Default()
	mockStorage := mocks.NewMockStorageInterface(t)

	userID := uuid.New()

	mockStorage.EXPECT().
		GetBalance(mock.Anything, userID).
		Return(nil, errors.New("database error")).
		Once()

	handler := NewBalanceHandler(mockStorage, logger)

	request := httptest.NewRequest(http.MethodGet, "/api/user/balance", nil)

	ctx := middleware.SetUserID(request.Context(), userID)
	request = request.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.GetBalance(w, request)

	res := w.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusInternalServerError, res.StatusCode)
}

func TestWithdraw_Success(t *testing.T) {
	logger := slog.Default()
	mockStorage := mocks.NewMockStorageInterface(t)

	userID := uuid.New()
	orderNumber := "79927398713"
	sum := 100.5

	mockStorage.EXPECT().
		CreateWithdrawal(mock.Anything, userID, orderNumber, sum).
		Return(nil).
		Once()

	handler := NewBalanceHandler(mockStorage, logger)

	reqBody := WithdrawRequest{
		Order: orderNumber,
		Sum:   sum,
	}
	body, _ := json.Marshal(reqBody)

	request := httptest.NewRequest(http.MethodPost, "/api/user/balance/withdraw", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")

	ctx := middleware.SetUserID(request.Context(), userID)
	request = request.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.Withdraw(w, request)

	res := w.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusOK, res.StatusCode)
}

func TestWithdraw_InvalidLuhn(t *testing.T) {
	logger := slog.Default()
	mockStorage := mocks.NewMockStorageInterface(t)
	handler := NewBalanceHandler(mockStorage, logger)

	userID := uuid.New()
	reqBody := WithdrawRequest{
		Order: "12345",
		Sum:   100.5,
	}
	body, _ := json.Marshal(reqBody)

	request := httptest.NewRequest(http.MethodPost, "/api/user/balance/withdraw", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")

	ctx := middleware.SetUserID(request.Context(), userID)
	request = request.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.Withdraw(w, request)

	res := w.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusUnprocessableEntity, res.StatusCode)
}

func TestGetWithdrawals_Success(t *testing.T) {
	logger := slog.Default()
	mockStorage := mocks.NewMockStorageInterface(t)

	userID := uuid.New()
	now := time.Now()

	withdrawal1 := models.NewWithdrawal("79927398713", 100.5, userID)
	withdrawal1.ProcessedAt = now

	withdrawals := []models.Withdrawal{*withdrawal1}

	mockStorage.EXPECT().
		GetWithdrawals(mock.Anything, userID).
		Return(withdrawals, nil).
		Once()

	handler := NewBalanceHandler(mockStorage, logger)

	request := httptest.NewRequest(http.MethodGet, "/api/user/withdrawals", nil)

	ctx := middleware.SetUserID(request.Context(), userID)
	request = request.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.GetWithdrawals(w, request)

	res := w.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusOK, res.StatusCode)
}

func TestGetWithdrawals_NoWithdrawals(t *testing.T) {
	logger := slog.Default()
	mockStorage := mocks.NewMockStorageInterface(t)

	userID := uuid.New()

	mockStorage.EXPECT().
		GetWithdrawals(mock.Anything, userID).
		Return([]models.Withdrawal{}, nil).
		Once()

	handler := NewBalanceHandler(mockStorage, logger)

	request := httptest.NewRequest(http.MethodGet, "/api/user/withdrawals", nil)

	ctx := middleware.SetUserID(request.Context(), userID)
	request = request.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.GetWithdrawals(w, request)

	res := w.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusNoContent, res.StatusCode)
}

func TestWithdraw_InvalidJSON(t *testing.T) {
	logger := slog.Default()
	mockStorage := mocks.NewMockStorageInterface(t)
	handler := NewBalanceHandler(mockStorage, logger)

	userID := uuid.New()

	request := httptest.NewRequest(http.MethodPost, "/api/user/balance/withdraw",
		bytes.NewReader([]byte("invalid json")))
	ctx := middleware.SetUserID(request.Context(), userID)
	request = request.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.Withdraw(w, request)

	res := w.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestWithdraw_InsufficientFunds(t *testing.T) {
	logger := slog.Default()
	mockStorage := mocks.NewMockStorageInterface(t)

	userID := uuid.New()
	orderNumber := "79927398713"
	sum := 1000.0

	mockStorage.EXPECT().
		CreateWithdrawal(mock.Anything, userID, orderNumber, sum).
		Return(storage.ErrInsufficientFunds).
		Once()

	handler := NewBalanceHandler(mockStorage, logger)

	reqBody := WithdrawRequest{
		Order: orderNumber,
		Sum:   sum,
	}
	body, _ := json.Marshal(reqBody)

	request := httptest.NewRequest(http.MethodPost, "/api/user/balance/withdraw", bytes.NewReader(body))
	ctx := middleware.SetUserID(request.Context(), userID)
	request = request.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.Withdraw(w, request)

	res := w.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusPaymentRequired, res.StatusCode)
}

func TestWithdraw_StorageError(t *testing.T) {
	logger := slog.Default()
	mockStorage := mocks.NewMockStorageInterface(t)

	userID := uuid.New()
	orderNumber := "79927398713"
	sum := 100.0

	mockStorage.EXPECT().
		CreateWithdrawal(mock.Anything, userID, orderNumber, sum).
		Return(errors.New("database error")).
		Once()

	handler := NewBalanceHandler(mockStorage, logger)

	reqBody := WithdrawRequest{
		Order: orderNumber,
		Sum:   sum,
	}
	body, _ := json.Marshal(reqBody)

	request := httptest.NewRequest(http.MethodPost, "/api/user/balance/withdraw", bytes.NewReader(body))
	ctx := middleware.SetUserID(request.Context(), userID)
	request = request.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.Withdraw(w, request)

	res := w.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusInternalServerError, res.StatusCode)
}

func TestGetWithdrawals_Error(t *testing.T) {
	logger := slog.Default()
	mockStorage := mocks.NewMockStorageInterface(t)

	userID := uuid.New()

	mockStorage.EXPECT().
		GetWithdrawals(mock.Anything, userID).
		Return(nil, errors.New("database error")).
		Once()

	handler := NewBalanceHandler(mockStorage, logger)

	request := httptest.NewRequest(http.MethodGet, "/api/user/withdrawals", nil)
	ctx := middleware.SetUserID(request.Context(), userID)
	request = request.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.GetWithdrawals(w, request)

	res := w.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusInternalServerError, res.StatusCode)
}

func TestGetBalance_DatabaseError(t *testing.T) {
	logger := slog.Default()
	mockStorage := mocks.NewMockStorageInterface(t)

	userID := uuid.New()
	mockStorage.EXPECT().
		GetBalance(mock.Anything, userID).
		Return(nil, errors.New("database error")).
		Once()

	handler := NewBalanceHandler(mockStorage, logger)

	request := httptest.NewRequest(http.MethodGet, "/api/user/balance", nil)
	ctx := middleware.SetUserID(request.Context(), userID)
	request = request.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.GetBalance(w, request)

	res := w.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusInternalServerError, res.StatusCode)
}
