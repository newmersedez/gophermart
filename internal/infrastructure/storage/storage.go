package storage

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"gophermart/internal/domain/models"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type StorageInterface interface {
	Close()
	CreateUser(ctx context.Context, login, passwordHash string) (uuid.UUID, error)
	GetUserByLogin(ctx context.Context, login string) (*models.User, error)
	CreateOrder(ctx context.Context, number string, userID uuid.UUID) error
	GetOrderByNumber(ctx context.Context, number string) (*models.Order, error)
	GetUserOrders(ctx context.Context, userID uuid.UUID) ([]models.Order, error)
	UpdateOrderStatus(ctx context.Context, number, status string, accrual *float64) error
	GetPendingOrders(ctx context.Context) ([]models.Order, error)
	GetBalance(ctx context.Context, userID uuid.UUID) (*models.Balance, error)
	CreateWithdrawal(ctx context.Context, userID uuid.UUID, order string, sum float64) error
	GetWithdrawals(ctx context.Context, userID uuid.UUID) ([]models.Withdrawal, error)
}

type Storage struct {
	pool   *pgxpool.Pool
	logger *slog.Logger
}

func NewStorage(ctx context.Context, dsn string, logger *slog.Logger) (*Storage, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	if err := runMigrations(dsn); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return &Storage{pool: pool, logger: logger}, nil
}

func runMigrations(dsn string) error {
	m, err := migrate.New("file://internal/infrastructure/storage/migrations", dsn)
	if err != nil {
		return fmt.Errorf("failed to create migration instance: %w", err)
	}
	defer func() {
		_, _ = m.Close()
	}()

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

func (s *Storage) Close() {
	s.pool.Close()
}

func (s *Storage) CreateUser(ctx context.Context, login, passwordHash string) (uuid.UUID, error) {
	userID := uuid.New()
	query := `INSERT INTO users(id, login, password_hash) VALUES($1, $2, $3)`

	_, err := s.pool.Exec(ctx, query, userID, login, passwordHash)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to create user '%s': %w", login, err)
	}

	return userID, nil
}

func (s *Storage) GetUserByLogin(ctx context.Context, login string) (*models.User, error) {
	query := `SELECT id, login, password_hash, created_at FROM users WHERE login = $1`

	user := models.NewUser("", "")
	err := s.pool.QueryRow(ctx, query, login).Scan(&user.ID, &user.Login, &user.PasswordHash, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user by login '%s': %w", login, err)
	}

	return user, nil
}

func (s *Storage) CreateOrder(ctx context.Context, number string, userID uuid.UUID) error {
	query := `INSERT INTO orders(number, user_id, status, uploaded_at) VALUES($1, $2, $3, NOW())`

	_, err := s.pool.Exec(ctx, query, number, userID, models.OrderStatusNew)
	if err != nil {
		return fmt.Errorf("failed to create order '%s' for user %s: %w", number, userID, err)
	}
	return nil
}

func (s *Storage) GetOrderByNumber(ctx context.Context, number string) (*models.Order, error) {
	query := `SELECT number, user_id, status, accrual, uploaded_at FROM orders WHERE number = $1`

	order := models.NewOrder("", uuid.Nil)
	err := s.pool.QueryRow(ctx, query, number).Scan(&order.Number, &order.UserID, &order.Status, &order.Accrual, &order.UploadedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get order by number '%s': %w", number, err)
	}

	return order, nil
}

func (s *Storage) GetUserOrders(ctx context.Context, userID uuid.UUID) ([]models.Order, error) {
	query := `SELECT number, user_id, status, accrual, uploaded_at FROM orders WHERE user_id = $1 ORDER BY uploaded_at DESC`

	rows, err := s.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query user orders for user %s: %w", userID, err)
	}
	defer rows.Close()

	var orders []models.Order
	for rows.Next() {
		order := models.NewOrder("", uuid.Nil)
		if err := rows.Scan(&order.Number, &order.UserID, &order.Status, &order.Accrual, &order.UploadedAt); err != nil {
			return nil, fmt.Errorf("failed to scan order row for user %s: %w", userID, err)
		}
		orders = append(orders, *order)
	}

	return orders, rows.Err()
}

func (s *Storage) UpdateOrderStatus(ctx context.Context, number, status string, accrual *float64) error {
	query := `UPDATE orders SET status = $1, accrual = $2 WHERE number = $3`

	_, err := s.pool.Exec(ctx, query, status, accrual, number)
	if err != nil {
		return fmt.Errorf("failed to update order '%s' status to '%s': %w", number, status, err)
	}
	return nil
}

func (s *Storage) GetPendingOrders(ctx context.Context) ([]models.Order, error) {
	query := `SELECT number, user_id, status, accrual, uploaded_at FROM orders WHERE status IN ($1, $2)`

	rows, err := s.pool.Query(ctx, query, models.OrderStatusNew, models.OrderStatusProcessing)
	if err != nil {
		return nil, fmt.Errorf("failed to query pending orders: %w", err)
	}
	defer rows.Close()

	var orders []models.Order
	for rows.Next() {
		order := models.NewOrder("", uuid.Nil)
		if err := rows.Scan(&order.Number, &order.UserID, &order.Status, &order.Accrual, &order.UploadedAt); err != nil {
			return nil, fmt.Errorf("failed to scan pending order row: %w", err)
		}
		orders = append(orders, *order)
	}

	return orders, rows.Err()
}

func (s *Storage) GetBalance(ctx context.Context, userID uuid.UUID) (*models.Balance, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction for user %s balance: %w", userID, err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var current float64
	err = tx.QueryRow(ctx, `SELECT COALESCE(SUM(accrual), 0) FROM orders WHERE user_id = $1 AND status = $2`, userID, models.OrderStatusProcessed).Scan(&current)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate current balance for user %s: %w", userID, err)
	}

	var withdrawn float64
	err = tx.QueryRow(ctx, `SELECT COALESCE(SUM(sum), 0) FROM withdrawals WHERE user_id = $1`, userID).Scan(&withdrawn)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate withdrawn amount for user %s: %w", userID, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit balance transaction for user %s: %w", userID, err)
	}

	balance := models.NewBalance(current-withdrawn, withdrawn)

	return balance, nil
}

func (s *Storage) CreateWithdrawal(ctx context.Context, userID uuid.UUID, order string, sum float64) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin withdrawal transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	balance, err := s.GetBalance(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get balance for withdrawal: %w", err)
	}

	if balance.Current < sum {
		return errors.New("insufficient funds")
	}

	query := `INSERT INTO withdrawals(order_number, sum, user_id, processed_at) VALUES($1, $2, $3, NOW())`
	_, err = tx.Exec(ctx, query, order, sum, userID)
	if err != nil {
		return fmt.Errorf("failed to insert withdrawal for order '%s': %w", order, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit withdrawal transaction: %w", err)
	}
	return nil
}

func (s *Storage) GetWithdrawals(ctx context.Context, userID uuid.UUID) ([]models.Withdrawal, error) {
	query := `SELECT order_number, sum, user_id, processed_at FROM withdrawals WHERE user_id = $1 ORDER BY processed_at DESC`

	rows, err := s.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query withdrawals for user %s: %w", userID, err)
	}
	defer rows.Close()

	var withdrawals []models.Withdrawal
	for rows.Next() {
		w := models.NewWithdrawal("", 0, uuid.Nil)
		if err := rows.Scan(&w.Order, &w.Sum, &w.UserID, &w.ProcessedAt); err != nil {
			return nil, fmt.Errorf("failed to scan withdrawal row for user %s: %w", userID, err)
		}
		withdrawals = append(withdrawals, *w)
	}

	return withdrawals, rows.Err()
}
