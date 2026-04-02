package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"gophermart/internal/middleware"
	"gophermart/internal/storage"
	"gophermart/internal/utils"

	"github.com/google/uuid"
)

type BalanceHandler struct {
	storage *storage.Storage
	logger  *slog.Logger
}

func NewBalanceHandler(storage *storage.Storage, logger *slog.Logger) *BalanceHandler {
	return &BalanceHandler{
		storage: storage,
		logger:  logger,
	}
}

func (h *BalanceHandler) GetBalance(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)

	balance, err := h.storage.GetBalance(r.Context(), userID)
	if err != nil {
		h.logger.Error("failed to get balance", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(balance)
}

type WithdrawRequest struct {
	Order string  `json:"order"`
	Sum   float64 `json:"sum"`
}

func (h *BalanceHandler) Withdraw(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)

	var req WithdrawRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	if !utils.ValidateLuhn(req.Order) {
		http.Error(w, "Invalid order number format", http.StatusUnprocessableEntity)
		return
	}

	err := h.storage.CreateWithdrawal(r.Context(), userID, req.Order, req.Sum)
	if err != nil {
		if err.Error() == "insufficient funds" {
			http.Error(w, "Insufficient funds", http.StatusPaymentRequired)
			return
		}
		h.logger.Error("failed to create withdrawal", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

type WithdrawalResponse struct {
	Order       string  `json:"order"`
	Sum         float64 `json:"sum"`
	ProcessedAt string  `json:"processed_at"`
}

func (h *BalanceHandler) GetWithdrawals(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)

	withdrawals, err := h.storage.GetWithdrawals(r.Context(), userID)
	if err != nil {
		h.logger.Error("failed to get withdrawals", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if len(withdrawals) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	var response []WithdrawalResponse
	for _, w := range withdrawals {
		response = append(response, WithdrawalResponse{
			Order:       w.Order,
			Sum:         w.Sum,
			ProcessedAt: w.ProcessedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
