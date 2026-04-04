package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"gophermart/internal/app/services/auth"
	"gophermart/internal/domain/models"
	"gophermart/internal/infrastructure/storage/mocks"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestRegister_Success(t *testing.T) {
	logger := slog.Default()
	mockStorage := mocks.NewMockStorageInterface(t)

	userID := uuid.New()
	mockStorage.EXPECT().
		CreateUser(mock.Anything, "testuser", mock.AnythingOfType("string")).
		Return(userID, nil).
		Once()

	handler := NewAuthHandler(mockStorage, logger)

	reqBody := RegisterRequest{
		Login:    "testuser",
		Password: "password123",
	}
	body, _ := json.Marshal(reqBody)

	request := httptest.NewRequest(http.MethodPost, "/api/user/register", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Register(w, request)

	res := w.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusOK, res.StatusCode)

	cookies := res.Cookies()
	require.Len(t, cookies, 1)
	assert.Equal(t, "token", cookies[0].Name)
	assert.NotEmpty(t, cookies[0].Value)
}

func TestRegister_UserAlreadyExists(t *testing.T) {
	logger := slog.Default()
	mockStorage := mocks.NewMockStorageInterface(t)

	pgErr := &pgconn.PgError{Code: "23505"}
	mockStorage.EXPECT().
		CreateUser(mock.Anything, "testuser", mock.AnythingOfType("string")).
		Return(uuid.Nil, pgErr).
		Once()

	handler := NewAuthHandler(mockStorage, logger)

	reqBody := RegisterRequest{
		Login:    "testuser",
		Password: "password123",
	}
	body, _ := json.Marshal(reqBody)

	request := httptest.NewRequest(http.MethodPost, "/api/user/register", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Register(w, request)

	res := w.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusConflict, res.StatusCode)
}

func TestLogin_Success(t *testing.T) {
	logger := slog.Default()
	mockStorage := mocks.NewMockStorageInterface(t)

	userID := uuid.New()
	passwordHash, _ := auth.HashPassword("password123")

	user := models.NewUser("testuser", passwordHash)
	user.ID = userID

	mockStorage.EXPECT().
		GetUserByLogin(mock.Anything, "testuser").
		Return(user, nil).
		Once()

	handler := NewAuthHandler(mockStorage, logger)

	reqBody := RegisterRequest{
		Login:    "testuser",
		Password: "password123",
	}
	body, _ := json.Marshal(reqBody)

	request := httptest.NewRequest(http.MethodPost, "/api/user/login", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Login(w, request)

	res := w.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusOK, res.StatusCode)

	cookies := res.Cookies()
	require.Len(t, cookies, 1)
	assert.Equal(t, "token", cookies[0].Name)
	assert.NotEmpty(t, cookies[0].Value)
}

func TestLogin_InvalidPassword(t *testing.T) {
	logger := slog.Default()
	mockStorage := mocks.NewMockStorageInterface(t)

	userID := uuid.New()
	passwordHash, _ := auth.HashPassword("correctpassword")

	user := models.NewUser("testuser", passwordHash)
	user.ID = userID

	mockStorage.EXPECT().
		GetUserByLogin(mock.Anything, "testuser").
		Return(user, nil).
		Once()

	handler := NewAuthHandler(mockStorage, logger)

	reqBody := RegisterRequest{
		Login:    "testuser",
		Password: "wrongpassword",
	}
	body, _ := json.Marshal(reqBody)

	request := httptest.NewRequest(http.MethodPost, "/api/user/login", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Login(w, request)

	res := w.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, res.StatusCode)
}

func TestRegister_InvalidJSON(t *testing.T) {
	logger := slog.Default()
	mockStorage := mocks.NewMockStorageInterface(t)
	handler := NewAuthHandler(mockStorage, logger)

	request := httptest.NewRequest(http.MethodPost, "/api/user/register",
		bytes.NewReader([]byte("invalid json")))
	w := httptest.NewRecorder()

	handler.Register(w, request)

	res := w.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestRegister_EmptyCredentials(t *testing.T) {
	logger := slog.Default()
	mockStorage := mocks.NewMockStorageInterface(t)
	handler := NewAuthHandler(mockStorage, logger)

	reqBody := RegisterRequest{
		Login:    "",
		Password: "",
	}
	body, _ := json.Marshal(reqBody)

	request := httptest.NewRequest(http.MethodPost, "/api/user/register", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.Register(w, request)

	res := w.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestRegister_StorageError(t *testing.T) {
	logger := slog.Default()
	mockStorage := mocks.NewMockStorageInterface(t)

	mockStorage.EXPECT().
		CreateUser(mock.Anything, "testuser", mock.Anything).
		Return(uuid.Nil, errors.New("database error")).
		Once()

	handler := NewAuthHandler(mockStorage, logger)

	reqBody := RegisterRequest{
		Login:    "testuser",
		Password: "password123",
	}
	body, _ := json.Marshal(reqBody)

	request := httptest.NewRequest(http.MethodPost, "/api/user/register", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.Register(w, request)

	res := w.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusInternalServerError, res.StatusCode)
}

func TestLogin_InvalidJSON(t *testing.T) {
	logger := slog.Default()
	mockStorage := mocks.NewMockStorageInterface(t)
	handler := NewAuthHandler(mockStorage, logger)

	request := httptest.NewRequest(http.MethodPost, "/api/user/login",
		bytes.NewReader([]byte("invalid json")))
	w := httptest.NewRecorder()

	handler.Login(w, request)

	res := w.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestLogin_EmptyCredentials(t *testing.T) {
	logger := slog.Default()
	mockStorage := mocks.NewMockStorageInterface(t)
	handler := NewAuthHandler(mockStorage, logger)

	reqBody := RegisterRequest{
		Login:    "",
		Password: "",
	}
	body, _ := json.Marshal(reqBody)

	request := httptest.NewRequest(http.MethodPost, "/api/user/login", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.Login(w, request)

	res := w.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestLogin_StorageError(t *testing.T) {
	logger := slog.Default()
	mockStorage := mocks.NewMockStorageInterface(t)

	mockStorage.EXPECT().
		GetUserByLogin(mock.Anything, "testuser").
		Return(nil, errors.New("database error")).
		Once()

	handler := NewAuthHandler(mockStorage, logger)

	reqBody := RegisterRequest{
		Login:    "testuser",
		Password: "password123",
	}
	body, _ := json.Marshal(reqBody)

	request := httptest.NewRequest(http.MethodPost, "/api/user/login", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.Login(w, request)

	res := w.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusInternalServerError, res.StatusCode)
}

func TestLogin_UserNotFound(t *testing.T) {
	logger := slog.Default()
	mockStorage := mocks.NewMockStorageInterface(t)

	mockStorage.EXPECT().
		GetUserByLogin(mock.Anything, "nonexistent").
		Return(nil, nil).
		Once()

	handler := NewAuthHandler(mockStorage, logger)

	reqBody := RegisterRequest{
		Login:    "nonexistent",
		Password: "password123",
	}
	body, _ := json.Marshal(reqBody)

	request := httptest.NewRequest(http.MethodPost, "/api/user/login", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.Login(w, request)

	res := w.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, res.StatusCode)
}

func TestRegister_CreateUserStorageError(t *testing.T) {
	logger := slog.Default()
	mockStorage := mocks.NewMockStorageInterface(t)

	mockStorage.EXPECT().
		CreateUser(mock.Anything, "testuser", mock.AnythingOfType("string")).
		Return(uuid.Nil, errors.New("general database error")).
		Once()

	handler := NewAuthHandler(mockStorage, logger)

	reqBody := RegisterRequest{
		Login:    "testuser",
		Password: "password123",
	}
	body, _ := json.Marshal(reqBody)

	request := httptest.NewRequest(http.MethodPost, "/api/user/register", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.Register(w, request)

	res := w.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusInternalServerError, res.StatusCode)
}
