package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID
	Login        string
	PasswordHash string
	CreatedAt    time.Time
}

type Order struct {
	Number     string
	UserID     uuid.UUID
	Status     string
	Accrual    *float64
	UploadedAt time.Time
}

type Withdrawal struct {
	Order       string
	Sum         float64
	UserID      uuid.UUID
	ProcessedAt time.Time
}

type Balance struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

func NewUser(login, passwordHash string) *User {
	return &User{
		ID:           uuid.New(),
		Login:        login,
		PasswordHash: passwordHash,
		CreatedAt:    time.Now(),
	}
}

func NewOrder(number string, userID uuid.UUID) *Order {
	return &Order{
		Number:     number,
		UserID:     userID,
		Status:     OrderStatusNew,
		UploadedAt: time.Now(),
	}
}

func NewWithdrawal(orderNumber string, sum float64, userID uuid.UUID) *Withdrawal {
	return &Withdrawal{
		Order:       orderNumber,
		Sum:         sum,
		UserID:      userID,
		ProcessedAt: time.Now(),
	}
}

func NewBalance(current, withdrawn float64) *Balance {
	return &Balance{
		Current:   current,
		Withdrawn: withdrawn,
	}
}

const (
	OrderStatusNew        = "NEW"
	OrderStatusProcessing = "PROCESSING"
	OrderStatusInvalid    = "INVALID"
	OrderStatusProcessed  = "PROCESSED"
)

const (
	AccrualStatusRegistered = "REGISTERED"
	AccrualStatusInvalid    = "INVALID"
	AccrualStatusProcessing = "PROCESSING"
	AccrualStatusProcessed  = "PROCESSED"
)
