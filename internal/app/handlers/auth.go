package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"gophermart/internal/app/services/auth"
	"gophermart/internal/domain/models"
	"gophermart/internal/infrastructure/storage"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

type AuthStorage interface {
	CreateUser(ctx context.Context, login, passwordHash string) (uuid.UUID, error)
	GetUserByLogin(ctx context.Context, login string) (*models.User, error)
}

type AuthHandler struct {
	storage AuthStorage
	logger  *slog.Logger
}

func NewAuthHandler(storage AuthStorage, logger *slog.Logger) *AuthHandler {
	return &AuthHandler{
		storage: storage,
		logger:  logger,
	}
}

type RegisterRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	if req.Login == "" || req.Password == "" {
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	passwordHash, err := auth.HashPassword(req.Password)
	if err != nil {
		h.logger.Error("failed to hash password", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	userID, err := h.storage.CreateUser(r.Context(), req.Login, passwordHash)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			http.Error(w, "Login already exists", http.StatusConflict)
			return
		}
		h.logger.Error("failed to create user", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	token, err := auth.GenerateToken(userID)
	if err != nil {
		h.logger.Error("failed to generate token", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
	})

	w.WriteHeader(http.StatusOK)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	if req.Login == "" || req.Password == "" {
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	user, err := h.storage.GetUserByLogin(r.Context(), req.Login)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}
		h.logger.Error("failed to get user", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if !auth.CheckPassword(req.Password, user.PasswordHash) {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	token, err := auth.GenerateToken(user.ID)
	if err != nil {
		h.logger.Error("failed to generate token", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
	})

	w.WriteHeader(http.StatusOK)
}
