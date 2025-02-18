package repositories

import (
	"context"
	"errors"
	"fmt"
	"loyalty/internal/app/apperrors"
	"loyalty/internal/app/entities"

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
	query := `
		UPDATE orders
		SET status_id = $1, attempts = $2, next_attempt = $3, accrual = $4, updated_at = NOW()
		WHERE order_number = $5
		RETURNING order_number, user_id, status_id
	`

	var updatedOrder entities.Order
	err := r.Pool.QueryRow(ctx, query, order.StatusID, order.Attempts, order.NextAttempt, order.Accrual, order.OrderID).
		Scan(&updatedOrder.OrderID, &updatedOrder.UserID, &updatedOrder.StatusID)

	if err != nil {
		return fmt.Errorf("failed to update order %d: %w", order.OrderID, err)
	}

	return nil
}

func (r *orderRepository) GetFreshOrders(ctx context.Context, limit int) ([]entities.Order, error) {
	query := `
		SELECT id, order_number, status_id, accrual, next_attempt, attempts, created_at, updated_at
		FROM orders
		WHERE status_id IN ($1, $2)
		  AND (next_attempt IS NULL OR next_attempt <= NOW())
		  AND attempts < $3
		ORDER BY created_at ASC
		LIMIT $4;
	`
	attempts := 3
	statusNew := entities.StatusNew
	statusProcessing := entities.StatusProcessing
	rows, err := r.Pool.Query(ctx, query, statusNew, statusProcessing, attempts, limit)
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
			&order.NextAttempt,
			&order.Attempts,
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
