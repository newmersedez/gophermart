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
