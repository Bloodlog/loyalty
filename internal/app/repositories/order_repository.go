package repositories

import (
	"context"
	"errors"
	"fmt"
	"loyalty/internal/app/apperrors"
	"loyalty/internal/app/entities"

	"github.com/jackc/pgx/v5"

	"github.com/jackc/pgx/v5/pgconn"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OrderRepositoryInterface interface {
	Store(ctx context.Context, order *entities.Order) error
	UpdateOrder(ctx context.Context, order *entities.Order) error
	GetFreshOrders(ctx context.Context, limit int) ([]entities.Order, error)
	GetByUserID(ctx context.Context, userID int) ([]entities.Order, error)
	GetTotalAccrualByUserID(ctx context.Context, userID int) (float64, error)
	GetByOrderNumber(ctx context.Context, orderNumber int64) (*entities.Order, error)
}

type orderRepository struct {
	Pool *pgxpool.Pool
}

func NewOrderRepository(db *pgxpool.Pool) OrderRepositoryInterface {
	return &orderRepository{
		Pool: db,
	}
}

func (r *orderRepository) GetByUserID(ctx context.Context, userID int) ([]entities.Order, error) {
	query := `SELECT order_number, status_id, accrual, updated_at
			FROM orders WHERE user_id = $1
			ORDER BY created_at ASC`
	rows, err := r.Pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get orders By UserID: %w", err)
	}
	defer rows.Close()

	var orders []entities.Order
	for rows.Next() {
		var order entities.Order
		err = rows.Scan(
			&order.OrderID,
			&order.StatusID,
			&order.Accrual,
			&order.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to get orders parse result %d: %w", order.OrderID, err)
		}
		orders = append(orders, order)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to get orders userID:%d: %w", userID, err)
	}

	return orders, nil
}

func (r *orderRepository) GetByOrderNumber(ctx context.Context, orderNumber int64) (*entities.Order, error) {
	query := `
		SELECT order_id, user_id, status_id
		FROM orders
		WHERE order_number = $1
	`

	var order entities.Order
	err := r.Pool.QueryRow(ctx, query, orderNumber).Scan(&order.OrderID, &order.UserID, &order.StatusID)
	if err != nil {
		return nil, apperrors.ErrOrderNotFound
	}

	return &order, nil
}

func (r *orderRepository) Store(ctx context.Context, order *entities.Order) error {
	query := `
		INSERT INTO orders (order_number, user_id, status_id)
		VALUES ($1, $2, $3)
	`

	_, err := r.Pool.Exec(ctx, query, order.OrderID, order.UserID, order.StatusID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
			err = apperrors.ErrDuplicateOrderID
		}
		return fmt.Errorf("failed to save order %d: %w", order.OrderID, err)
	}

	return nil
}

func (r *orderRepository) UpdateOrder(ctx context.Context, order *entities.Order) error {
	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer func(tx pgx.Tx, ctx context.Context) {
		_ = tx.Rollback(ctx)
	}(tx, ctx)

	query := `
		UPDATE orders
		SET status_id = $1, accrual = $2, updated_at = NOW()
		WHERE order_number = $3
		RETURNING order_number, user_id, status_id
	`

	var updatedOrder entities.Order
	err = tx.QueryRow(ctx, query, order.StatusID, order.Accrual, order.OrderID).
		Scan(&updatedOrder.OrderID, &updatedOrder.UserID, &updatedOrder.StatusID)

	if err != nil {
		return fmt.Errorf("failed to update order %d: %w", order.OrderID, err)
	}

	if order.Accrual.Valid && order.Accrual.Float64 > 0 {
		queryUser := `
			UPDATE users
			SET balance = balance + $1
			WHERE user_id = $2
		`
		_, err = tx.Exec(ctx, queryUser, updatedOrder.Accrual, updatedOrder.UserID)
		if err != nil {
			return fmt.Errorf("failed to update user balance for user %d: %w", updatedOrder.UserID, err)
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (r *orderRepository) GetFreshOrders(ctx context.Context, limit int) ([]entities.Order, error) {
	query := `
		SELECT id, order_number, status_id, accrual, next_attempt, attempts, created_at, updated_at
		FROM orders
		WHERE status_id IN ($1, $2)
		ORDER BY created_at ASC
		LIMIT $3;
	`
	statusNew := entities.StatusNew
	statusProcessing := entities.StatusProcessing
	rows, err := r.Pool.Query(ctx, query, statusNew, statusProcessing, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get Fresh Orders: %w", err)
	}
	defer rows.Close()

	var orders []entities.Order
	for rows.Next() {
		var order entities.Order
		err = rows.Scan(
			&order.ID,
			&order.OrderID,
			&order.StatusID,
			&order.Accrual,
			&order.CreatedAt,
			&order.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to Scan orders: %w", err)
		}
		orders = append(orders, order)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to get orders: %w", err)
	}

	return orders, nil
}

func (r *orderRepository) GetTotalAccrualByUserID(ctx context.Context, userID int) (float64, error) {
	query := `
		SELECT COALESCE(SUM(accrual), 0)
		FROM orders
		WHERE user_id = $1
	`

	var totalAccrual float64
	err := r.Pool.QueryRow(ctx, query, userID).Scan(&totalAccrual)
	if err != nil {
		return 0, fmt.Errorf("failed to get total accrual for user %d: %w", userID, err)
	}

	return totalAccrual, nil
}
