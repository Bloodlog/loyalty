package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"gophermart/internal/app/apperrors"
	"gophermart/internal/app/entities"

	"github.com/jackc/pgx/v5"

	"github.com/jackc/pgx/v5/pgconn"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgxpool"
)

type WithdrawRepositoryInterface interface {
	StoreAndUpdateBalance(ctx context.Context, withdraw entities.Withdraw) error
	GetTotalWithdrawByUserID(ctx context.Context, userID int) (float64, error)
	GetByUserID(ctx context.Context, userID int) ([]entities.Withdraw, error)
}

type withdrawRepository struct {
	Pool *pgxpool.Pool
}

func NewWithdrawRepository(db *pgxpool.Pool) WithdrawRepositoryInterface {
	return &withdrawRepository{
		Pool: db,
	}
}

func (r *withdrawRepository) StoreAndUpdateBalance(ctx context.Context, withdraw entities.Withdraw) error {
	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer func(tx pgx.Tx, ctx context.Context) {
		_ = tx.Rollback(ctx)
	}(tx, ctx)

	query := `
		INSERT INTO withdraws (order_number, user_id, withdraw)
		VALUES ($1, $2, $3)
	`
	_, err = tx.Exec(ctx, query, withdraw.OrderID, withdraw.UserID, withdraw.Withdraw)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
			err = apperrors.ErrDuplicateOrderID
		}
		return fmt.Errorf("failed to save withdraw: %w", err)
	}

	if withdraw.Withdraw > 0 {
		queryGetBalance := `
	        SELECT COALESCE(balance, 0)
	        FROM users
	        WHERE id = $1
	    `
		var currentBalance sql.NullFloat64
		err = tx.QueryRow(ctx, queryGetBalance, withdraw.UserID).Scan(&currentBalance)
		if err != nil {
			return fmt.Errorf("failed to get current balance for user %d: %w", withdraw.UserID, err)
		}
		newBalance := currentBalance.Float64 - withdraw.Withdraw
		queryUser := `
			UPDATE users
			SET balance = $1
			WHERE id = $2
		`
		_, err = tx.Exec(ctx, queryUser, newBalance, withdraw.UserID)
		if err != nil {
			return fmt.Errorf(
				"failed to update user balance for user %d: %w",
				withdraw.UserID,
				err,
			)
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (r *withdrawRepository) GetTotalWithdrawByUserID(ctx context.Context, userID int) (float64, error) {
	query := `
		SELECT COALESCE(SUM(withdraw), 0)
		FROM withdraws
		WHERE user_id = $1
	`

	var totalAccrual float64
	err := r.Pool.QueryRow(ctx, query, userID).Scan(&totalAccrual)
	if err != nil {
		return 0, fmt.Errorf("failed to get total accrual for user %d: %w", userID, err)
	}

	return totalAccrual, nil
}

func (r *withdrawRepository) GetByUserID(ctx context.Context, userID int) ([]entities.Withdraw, error) {
	query := `
		SELECT order_number, withdraw, created_at
		FROM withdraws
		WHERE user_id = $1
		ORDER BY created_at ASC
	`
	rows, err := r.Pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get withdraw: %w", err)
	}
	defer rows.Close()

	var withdraws []entities.Withdraw
	for rows.Next() {
		var withdraw entities.Withdraw
		err = rows.Scan(
			&withdraw.OrderID,
			&withdraw.Withdraw,
			&withdraw.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to get withdraw: %w", err)
		}
		withdraws = append(withdraws, withdraw)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to get withdraw: %w", err)
	}

	return withdraws, nil
}
