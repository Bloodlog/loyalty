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

type WithdrawRepositoryInterface interface {
	Store(ctx context.Context, withdraw entities.Withdraw) error
	GetTotalWithdrawByUserID(ctx context.Context, userID int) (float64, error)
	GetByUserID(ctx context.Context, userID int) ([]entities.Withdraw, error)
}

type withdrawRepository struct {
	Pool *pgxpool.Pool
}

func NewWithdrawRepository(DB *pgxpool.Pool) WithdrawRepositoryInterface {
	return &withdrawRepository{
		Pool: DB,
	}
}

func (r *withdrawRepository) Store(ctx context.Context, withdraw entities.Withdraw) error {
	query := `
		INSERT INTO withdraws (order_number, user_id, withdraw)
		VALUES ($1, $2, $3)
	`

	_, err := r.Pool.Exec(ctx, query, withdraw.OrderId, withdraw.UserID, withdraw.Withdraw)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
			err = apperrors.ErrDuplicateOrderID
		}
		return err
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
	query := `SELECT order_number, withdraw, created_at
			FROM withdraws WHERE user_id = $1
			ORDER BY created_at ASC`
	rows, err := r.Pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var withdraws []entities.Withdraw
	for rows.Next() {
		var withdraw entities.Withdraw
		err := rows.Scan(
			&withdraw.OrderId,
			&withdraw.Withdraw,
			&withdraw.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		withdraws = append(withdraws, withdraw)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return withdraws, nil
}
