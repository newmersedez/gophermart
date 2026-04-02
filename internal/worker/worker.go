package worker

import (
	"context"
	"log/slog"
	"time"

	"gophermart/internal/accrual"
	"gophermart/internal/models"
	"gophermart/internal/storage"
)

type Worker struct {
	storage       *storage.Storage
	accrualClient *accrual.Client
	logger        *slog.Logger
	interval      time.Duration
}

func NewWorker(storage *storage.Storage, accrualClient *accrual.Client, logger *slog.Logger) *Worker {
	return &Worker{
		storage:       storage,
		accrualClient: accrualClient,
		logger:        logger,
		interval:      5 * time.Second,
	}
}

func (w *Worker) Start(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.processOrders(ctx)
		}
	}
}

func (w *Worker) processOrders(ctx context.Context) {
	orders, err := w.storage.GetPendingOrders(ctx)
	if err != nil {
		w.logger.Error("failed to get pending orders", "error", err)
		return
	}

	for _, order := range orders {
		select {
		case <-ctx.Done():
			return
		default:
			w.processOrder(ctx, order)
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func (w *Worker) processOrder(ctx context.Context, order models.Order) {
	resp, err := w.accrualClient.GetOrderAccrual(ctx, order.Number)
	if err != nil {
		if err.Error() == "too many requests" {
			time.Sleep(5 * time.Second)
		}
		w.logger.Debug("failed to get accrual for order", "order", order.Number, "error", err)
		return
	}

	if resp == nil {
		return
	}

	status := w.accrualClient.MapStatus(resp.Status)
	if err := w.storage.UpdateOrderStatus(ctx, order.Number, status, resp.Accrual); err != nil {
		w.logger.Error("failed to update order status", "order", order.Number, "error", err)
	}
}
