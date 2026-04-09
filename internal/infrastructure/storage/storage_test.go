package storage

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func getTestDSN() string {
	dsn := os.Getenv("TEST_DATABASE_URI")
	if dsn == "" {
		dsn = "postgres://postgres:test@localhost:5432/gophermart_test?sslmode=disable"
	}
	return dsn
}

func setupTestStorage(t *testing.T) *Storage {
	dsn := getTestDSN()
	logger := slog.Default()
	ctx := context.Background()

	storage, err := NewStorage(ctx, dsn, logger)
	if err != nil {
		t.Skip("Database not available, skipping storage tests:", err)
	}

	_, _ = storage.pool.Exec(ctx, "TRUNCATE users, orders, withdrawals CASCADE")

	return storage
}

func TestNewStorage(t *testing.T) {
	dsn := getTestDSN()
	logger := slog.Default()
	ctx := context.Background()

	storage, err := NewStorage(ctx, dsn, logger)
	if err != nil {
		t.Skip("Database not available:", err)
	}
	defer storage.Close()

	assert.NotNil(t, storage)
	assert.NotNil(t, storage.pool)
}

func TestCreateUser(t *testing.T) {
	storage := setupTestStorage(t)
	defer storage.Close()

	ctx := context.Background()
	login := "testuser"
	password := "hashedpassword"

	userID, err := storage.CreateUser(ctx, login, password)

	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, userID)
}

func TestGetUserByLogin(t *testing.T) {
	storage := setupTestStorage(t)
	defer storage.Close()

	ctx := context.Background()
	login := "testuser"
	password := "hashedpassword"

	userID, err := storage.CreateUser(ctx, login, password)
	require.NoError(t, err)

	user, err := storage.GetUserByLogin(ctx, login)
	require.NoError(t, err)
	require.NotNil(t, user)

	assert.Equal(t, userID, user.ID)
	assert.Equal(t, login, user.Login)
}

func TestGetUserByLogin_NotFound(t *testing.T) {
	storage := setupTestStorage(t)
	defer storage.Close()

	ctx := context.Background()

	user, err := storage.GetUserByLogin(ctx, "nonexistent")
	require.ErrorIs(t, err, storage.ErrUserNotFound)
	assert.Nil(t, user)
}

func TestCreateOrder(t *testing.T) {
	storage := setupTestStorage(t)
	defer storage.Close()

	ctx := context.Background()
	userID, _ := storage.CreateUser(ctx, "testuser", "pass")

	err := storage.CreateOrder(ctx, "12345678903", userID)
	assert.NoError(t, err)
}

func TestGetOrderByNumber(t *testing.T) {
	storage := setupTestStorage(t)
	defer storage.Close()

	ctx := context.Background()
	userID, _ := storage.CreateUser(ctx, "testuser", "pass")
	orderNumber := "12345678903"

	storage.CreateOrder(ctx, orderNumber, userID)

	order, err := storage.GetOrderByNumber(ctx, orderNumber)
	require.NoError(t, err)
	require.NotNil(t, order)

	assert.Equal(t, orderNumber, order.Number)
	assert.Equal(t, userID, order.UserID)
}

func TestGetUserOrders(t *testing.T) {
	storage := setupTestStorage(t)
	defer storage.Close()

	ctx := context.Background()
	userID, _ := storage.CreateUser(ctx, "testuser", "pass")

	storage.CreateOrder(ctx, "12345678903", userID)
	storage.CreateOrder(ctx, "79927398713", userID)

	orders, err := storage.GetUserOrders(ctx, userID)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(orders), 2)
}

func TestNewStorage_InvalidDSN(t *testing.T) {
	logger := slog.Default()
	ctx := context.Background()

	_, err := NewStorage(ctx, "invalid://dsn", logger)
	assert.Error(t, err)
}

func TestNewStorage_CannotConnect(t *testing.T) {
	logger := slog.Default()
	ctx := context.Background()

	_, err := NewStorage(ctx, "postgres://invalid", logger)
	assert.Error(t, err)
}

func TestClose(t *testing.T) {
	storage := setupTestStorage(t)

	assert.NotPanics(t, func() {
		storage.Close()
	})
}

func TestUpdateOrderStatus(t *testing.T) {
	storage := setupTestStorage(t)
	defer storage.Close()

	ctx := context.Background()
	userID, _ := storage.CreateUser(ctx, "testuser", "pass")
	orderNumber := "12345678903"

	storage.CreateOrder(ctx, orderNumber, userID)

	accrual := 100.5
	err := storage.UpdateOrderStatus(ctx, orderNumber, "PROCESSED", &accrual)
	require.NoError(t, err)

	order, err := storage.GetOrderByNumber(ctx, orderNumber)
	require.NoError(t, err)
	require.NotNil(t, order)

	assert.Equal(t, "PROCESSED", order.Status)
	assert.NotNil(t, order.Accrual)
	assert.InEpsilon(t, accrual, *order.Accrual, 1e-9)
}

func TestGetPendingOrders(t *testing.T) {
	storage := setupTestStorage(t)
	defer storage.Close()

	ctx := context.Background()
	userID, _ := storage.CreateUser(ctx, "testuser", "pass")

	storage.CreateOrder(ctx, "12345678903", userID)
	storage.CreateOrder(ctx, "79927398713", userID)

	accrual := 50.0
	storage.UpdateOrderStatus(ctx, "12345678903", "PROCESSED", &accrual)

	pendingOrders, err := storage.GetPendingOrders(ctx)
	require.NoError(t, err)

	found := false
	for _, order := range pendingOrders {
		if order.Number == "79927398713" && order.Status == "NEW" {
			found = true
			break
		}
	}
	assert.True(t, found)
}

func TestGetBalance(t *testing.T) {
	storage := setupTestStorage(t)
	defer storage.Close()

	ctx := context.Background()
	userID, _ := storage.CreateUser(ctx, "testuser", "pass")

	storage.CreateOrder(ctx, "12345678903", userID)
	accrual := 150.0
	storage.UpdateOrderStatus(ctx, "12345678903", "PROCESSED", &accrual)

	balance, err := storage.GetBalance(ctx, userID)
	require.NoError(t, err)
	require.NotNil(t, balance)

	assert.InEpsilon(t, accrual, balance.Current, 1e-9)
	assert.InDelta(t, 0.0, balance.Withdrawn, 1e-9)
}

func TestCreateWithdrawal(t *testing.T) {
	storage := setupTestStorage(t)
	defer storage.Close()

	ctx := context.Background()
	userID, _ := storage.CreateUser(ctx, "testuser", "pass")

	storage.CreateOrder(ctx, "12345678903", userID)
	accrual := 200.0
	storage.UpdateOrderStatus(ctx, "12345678903", "PROCESSED", &accrual)

	err := storage.CreateWithdrawal(ctx, userID, "2377225624", 100.0)
	require.NoError(t, err)

	balance, err := storage.GetBalance(ctx, userID)
	require.NoError(t, err)
	assert.InEpsilon(t, 100.0, balance.Current, 1e-9)
	assert.InEpsilon(t, 100.0, balance.Withdrawn, 1e-9)
}

func TestCreateWithdrawal_InsufficientFunds(t *testing.T) {
	storage := setupTestStorage(t)
	defer storage.Close()

	ctx := context.Background()
	userID, _ := storage.CreateUser(ctx, "testuser", "pass")

	err := storage.CreateWithdrawal(ctx, userID, "2377225624", 100.0)
	assert.Error(t, err)
	assert.ErrorContains(t, err, "insufficient funds")
}

func TestGetWithdrawals(t *testing.T) {
	storage := setupTestStorage(t)
	defer storage.Close()

	ctx := context.Background()
	userID, _ := storage.CreateUser(ctx, "testuser", "pass")

	storage.CreateOrder(ctx, "12345678903", userID)
	accrual := 200.0
	storage.UpdateOrderStatus(ctx, "12345678903", "PROCESSED", &accrual)

	storage.CreateWithdrawal(ctx, userID, "2377225624", 50.0)
	storage.CreateWithdrawal(ctx, userID, "4000000000000002", 30.0)

	withdrawals, err := storage.GetWithdrawals(ctx, userID)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(withdrawals), 2)

	if len(withdrawals) >= 2 {
		assert.True(t, withdrawals[0].ProcessedAt.After(withdrawals[1].ProcessedAt) ||
			withdrawals[0].ProcessedAt.Equal(withdrawals[1].ProcessedAt))
	}
}
