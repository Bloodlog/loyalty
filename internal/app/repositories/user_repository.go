package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"gophermart/internal/app/apperrors"
	"gophermart/internal/app/entities"

	"github.com/jackc/pgx/v5/pgconn"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepositoryInterface interface {
	GetByLogin(ctx context.Context, login string) (entities.User, error)
	Store(ctx context.Context, user entities.User) (entities.User, error)
	IsExistByID(ctx context.Context, id int) bool
	GetBalanceByUserID(ctx context.Context, userID int) (float64, error)
}

type userRepository struct {
	Pool *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) UserRepositoryInterface {
	return &userRepository{
		Pool: db,
	}
}

func (r *userRepository) GetByLogin(ctx context.Context, login string) (entities.User, error) {
	var user entities.User
	query := `
		SELECT id, login, password
		FROM users
		WHERE login = $1
	`
	err := r.Pool.QueryRow(ctx, query, login).Scan(&user.ID, &user.Login, &user.Password)
	if err != nil {
		return user, fmt.Errorf("failed to get login: %w", err)
	}
	return user, nil
}

func (r *userRepository) Store(ctx context.Context, user entities.User) (entities.User, error) {
	query := `
		INSERT INTO users (login, password)
		VALUES ($1, $2)
		RETURNING id
	`
	err := r.Pool.QueryRow(ctx, query, user.Login, user.Password).Scan(&user.ID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
			err = apperrors.ErrDuplicateLogin
		}
		return user, fmt.Errorf("failed to save user: %w", err)
	}

	return user, nil
}

func (r *userRepository) IsExistByID(ctx context.Context, id int) bool {
	var exists bool
	query := `
		SELECT EXISTS (
			SELECT 1
			FROM users
			WHERE id = $1
		)
	`
	err := r.Pool.QueryRow(ctx, query, id).Scan(&exists)
	if err != nil {
		return false
	}
	return exists
}

func (r *userRepository) GetBalanceByUserID(ctx context.Context, userID int) (float64, error) {
	query := `
		SELECT balance
		FROM users
		WHERE id = $1
	`

	var totalAccrual sql.NullFloat64
	err := r.Pool.QueryRow(ctx, query, userID).Scan(&totalAccrual)
	if err != nil {
		return 0, fmt.Errorf("failed to get total accrual for user %d: %w", userID, err)
	}

	if !totalAccrual.Valid {
		return 0, nil
	}

	return totalAccrual.Float64, nil
}
