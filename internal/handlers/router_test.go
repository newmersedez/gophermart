package handlers

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"gophermart/internal/storage/mocks"

	"github.com/stretchr/testify/assert"
)

func TestNewRouter(t *testing.T) {
	logger := slog.Default()
	mockStorage := mocks.NewMockStorageInterface(t)

	router := NewRouter(mockStorage, logger)

	assert.NotNil(t, router)
	assert.NotNil(t, router.authHandler)
	assert.NotNil(t, router.orderHandler)
	assert.NotNil(t, router.balanceHandler)
}

func TestRoutes(t *testing.T) {
	logger := slog.Default()
	mockStorage := mocks.NewMockStorageInterface(t)

	router := NewRouter(mockStorage, logger)
	handler := router.Routes(logger)

	assert.NotNil(t, handler)

	// Test that routes are registered properly
	server := httptest.NewServer(handler)
	defer server.Close()

	// Test unprotected routes exist (should not be 404)
	resp, err := http.Post(server.URL+"/api/user/register", "application/json", nil)
	assert.NoError(t, err)
	defer resp.Body.Close()
	assert.NotEqual(t, http.StatusNotFound, resp.StatusCode)

	resp, err = http.Post(server.URL+"/api/user/login", "application/json", nil)
	assert.NoError(t, err)
	defer resp.Body.Close()
	assert.NotEqual(t, http.StatusNotFound, resp.StatusCode)

	// Test protected routes exist (will get 401 without auth)
	resp, err = http.Get(server.URL + "/api/user/orders")
	assert.NoError(t, err)
	defer resp.Body.Close()
	// Should be 401 (unauthorized) not 404 (not found)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	resp, err = http.Get(server.URL + "/api/user/balance")
	assert.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}
