package models

import (
	"testing"

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
