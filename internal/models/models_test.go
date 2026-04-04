package models

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestOrderStatusConstants(t *testing.T) {
	assert.Equal(t, "NEW", OrderStatusNew)
	assert.Equal(t, "PROCESSING", OrderStatusProcessing)
	assert.Equal(t, "INVALID", OrderStatusInvalid)
	assert.Equal(t, "PROCESSED", OrderStatusProcessed)
}

func TestAccrualStatusConstants(t *testing.T) {
	assert.Equal(t, "REGISTERED", AccrualStatusRegistered)
	assert.Equal(t, "INVALID", AccrualStatusInvalid)
	assert.Equal(t, "PROCESSING", AccrualStatusProcessing)
	assert.Equal(t, "PROCESSED", AccrualStatusProcessed)
}

func TestNewUser(t *testing.T) {
	login := "testuser"
	passwordHash := "hash123"

	user := NewUser(login, passwordHash)

	assert.NotNil(t, user)
	assert.NotEqual(t, uuid.Nil, user.ID)
	assert.Equal(t, login, user.Login)
	assert.Equal(t, passwordHash, user.PasswordHash)
	assert.False(t, user.CreatedAt.IsZero())
}

func TestNewOrder(t *testing.T) {
	number := "12345678903"
	userID := uuid.New()

	order := NewOrder(number, userID)

	assert.NotNil(t, order)
	assert.Equal(t, number, order.Number)
	assert.Equal(t, userID, order.UserID)
	assert.Equal(t, OrderStatusNew, order.Status)
	assert.Nil(t, order.Accrual)
	assert.False(t, order.UploadedAt.IsZero())
}

func TestNewWithdrawal(t *testing.T) {
	orderNumber := "79927398713"
	sum := 100.5
	userID := uuid.New()

	withdrawal := NewWithdrawal(orderNumber, sum, userID)

	assert.NotNil(t, withdrawal)
	assert.Equal(t, orderNumber, withdrawal.Order)
	assert.Equal(t, sum, withdrawal.Sum)
	assert.Equal(t, userID, withdrawal.UserID)
	assert.False(t, withdrawal.ProcessedAt.IsZero())
}

func TestNewBalance(t *testing.T) {
	current := 500.75
	withdrawn := 100.25

	balance := NewBalance(current, withdrawn)

	assert.NotNil(t, balance)
	assert.Equal(t, current, balance.Current)
	assert.Equal(t, withdrawn, balance.Withdrawn)
}
