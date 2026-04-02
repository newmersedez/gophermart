package storage

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"gophermart/internal/models"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

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
	m, err := migrate.New("file://internal/storage/migrations", dsn)
	if err != nil {
		return err
	}
	defer func() { 
		_, _ = m.Close() 
	}()

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
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
		return uuid.Nil, err
	}

	return userID, nil
}

func (s *Storage) GetUserByLogin(ctx context.Context, login string) (*models.User, error) {
	query := `SELECT id, login, password_hash, created_at FROM users WHERE login = $1`

	var user models.User
	err := s.pool.QueryRow(ctx, query, login).Scan(&user.ID, &user.Login, &user.PasswordHash, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &user, nil
}

func (s *Storage) CreateOrder(ctx context.Context, number string, userID uuid.UUID) error {
	query := `INSERT INTO orders(number, user_id, status, uploaded_at) VALUES($1, $2, $3, NOW())`

	_, err := s.pool.Exec(ctx, query, number, userID, models.OrderStatusNew)
	return err
}

func (s *Storage) GetOrderByNumber(ctx context.Context, number string) (*models.Order, error) {
	query := `SELECT number, user_id, status, accrual, uploaded_at FROM orders WHERE number = $1`

	var order models.Order
	err := s.pool.QueryRow(ctx, query, number).Scan(&order.Number, &order.UserID, &order.Status, &order.Accrual, &order.UploadedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &order, nil
}

func (s *Storage) GetUserOrders(ctx context.Context, userID uuid.UUID) ([]models.Order, error) {
	query := `SELECT number, user_id, status, accrual, uploaded_at FROM orders WHERE user_id = $1 ORDER BY uploaded_at DESC`

	rows, err := s.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []models.Order
	for rows.Next() {
		var order models.Order
		if err := rows.Scan(&order.Number, &order.UserID, &order.Status, &order.Accrual, &order.UploadedAt); err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}

	return orders, rows.Err()
}

func (s *Storage) UpdateOrderStatus(ctx context.Context, number, status string, accrual *float64) error {
	query := `UPDATE orders SET status = $1, accrual = $2 WHERE number = $3`

	_, err := s.pool.Exec(ctx, query, status, accrual, number)
	return err
}

func (s *Storage) GetPendingOrders(ctx context.Context) ([]models.Order, error) {
	query := `SELECT number, user_id, status, accrual, uploaded_at FROM orders WHERE status IN ($1, $2)`

	rows, err := s.pool.Query(ctx, query, models.OrderStatusNew, models.OrderStatusProcessing)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []models.Order
	for rows.Next() {
		var order models.Order
		if err := rows.Scan(&order.Number, &order.UserID, &order.Status, &order.Accrual, &order.UploadedAt); err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}

	return orders, rows.Err()
}

func (s *Storage) GetBalance(ctx context.Context, userID uuid.UUID) (*models.Balance, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var current float64
	err = tx.QueryRow(ctx, `SELECT COALESCE(SUM(accrual), 0) FROM orders WHERE user_id = $1 AND status = $2`, userID, models.OrderStatusProcessed).Scan(&current)
	if err != nil {
		return nil, err
	}

	var withdrawn float64
	err = tx.QueryRow(ctx, `SELECT COALESCE(SUM(sum), 0) FROM withdrawals WHERE user_id = $1`, userID).Scan(&withdrawn)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &models.Balance{
		Current:   current - withdrawn,
		Withdrawn: withdrawn,
	}, nil
}

func (s *Storage) CreateWithdrawal(ctx context.Context, userID uuid.UUID, order string, sum float64) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	balance, err := s.GetBalance(ctx, userID)
	if err != nil {
		return err
	}

	if balance.Current < sum {
		return errors.New("insufficient funds")
	}

	query := `INSERT INTO withdrawals(order_number, sum, user_id, processed_at) VALUES($1, $2, $3, NOW())`
	_, err = tx.Exec(ctx, query, order, sum, userID)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (s *Storage) GetWithdrawals(ctx context.Context, userID uuid.UUID) ([]models.Withdrawal, error) {
	query := `SELECT order_number, sum, user_id, processed_at FROM withdrawals WHERE user_id = $1 ORDER BY processed_at DESC`

	rows, err := s.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var withdrawals []models.Withdrawal
	for rows.Next() {
		var w models.Withdrawal
		if err := rows.Scan(&w.Order, &w.Sum, &w.UserID, &w.ProcessedAt); err != nil {
			return nil, err
		}
		withdrawals = append(withdrawals, w)
	}

	return withdrawals, rows.Err()
}
