package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"gophermart/internal/middleware"
	"gophermart/internal/storage"
	"gophermart/internal/utils"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

type OrderHandler struct {
	storage storage.StorageInterface
	logger  *slog.Logger
}

func NewOrderHandler(storage storage.StorageInterface, logger *slog.Logger) *OrderHandler {
	return &OrderHandler{
		storage: storage,
		logger:  logger,
	}
}

func (h *OrderHandler) UploadOrder(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	orderNumber := string(body)
	if orderNumber == "" {
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	if !utils.ValidateLuhn(orderNumber) {
		http.Error(w, "Invalid order number format", http.StatusUnprocessableEntity)
		return
	}

	existingOrder, err := h.storage.GetOrderByNumber(r.Context(), orderNumber)
	if err != nil {
		h.logger.Error("failed to get order", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if existingOrder != nil {
		if existingOrder.UserID == userID {
			w.WriteHeader(http.StatusOK)
			return
		}
		http.Error(w, "Order already uploaded by another user", http.StatusConflict)
		return
	}

	err = h.storage.CreateOrder(r.Context(), orderNumber, userID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			http.Error(w, "Order already uploaded by another user", http.StatusConflict)
			return
		}
		h.logger.Error("failed to create order", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

type OrderResponse struct {
	Number     string   `json:"number"`
	Status     string   `json:"status"`
	Accrual    *float64 `json:"accrual,omitempty"`
	UploadedAt string   `json:"uploaded_at"`
}

func (h *OrderHandler) GetOrders(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)

	orders, err := h.storage.GetUserOrders(r.Context(), userID)
	if err != nil {
		h.logger.Error("failed to get orders", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if len(orders) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	var response []OrderResponse
	for _, order := range orders {
		response = append(response, OrderResponse{
			Number:     order.Number,
			Status:     order.Status,
			Accrual:    order.Accrual,
			UploadedAt: order.UploadedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("failed to encode response", "error", err)
	}
}
