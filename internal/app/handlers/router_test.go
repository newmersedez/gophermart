package handlers

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"gophermart/internal/infrastructure/storage/mocks"

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

	
	server := httptest.NewServer(handler)
	defer server.Close()

	
	resp, err := http.Post(server.URL+"/api/user/register", "application/json", nil)
	assert.NoError(t, err)
	defer resp.Body.Close()
	assert.NotEqual(t, http.StatusNotFound, resp.StatusCode)

	resp, err = http.Post(server.URL+"/api/user/login", "application/json", nil)
	assert.NoError(t, err)
	defer resp.Body.Close()
	assert.NotEqual(t, http.StatusNotFound, resp.StatusCode)

	
	resp, err = http.Get(server.URL + "/api/user/orders")
	assert.NoError(t, err)
	defer resp.Body.Close()
	
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	resp, err = http.Get(server.URL + "/api/user/balance")
	assert.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}
