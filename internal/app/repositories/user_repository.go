package repositories

import (
	"context"
	"errors"
	"loyalty/internal/app/apperrors"
	"loyalty/internal/app/entities"

	"github.com/jackc/pgx/v5/pgconn"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepositoryInterface interface {
	GetByLogin(ctx context.Context, login string) (entities.User, error)
	Store(ctx context.Context, user entities.User) (entities.User, error)
	IsExistById(ctx context.Context, id int) bool
}

type userRepository struct {
	Pool *pgxpool.Pool
}

func NewUserRepository(DB *pgxpool.Pool) UserRepositoryInterface {
	return &userRepository{
		Pool: DB,
	}
}

func (r *userRepository) GetByLogin(ctx context.Context, login string) (entities.User, error) {
	var user entities.User
	query := "SELECT id, login, password FROM users WHERE login = $1"
	err := r.Pool.QueryRow(ctx, query, login).Scan(&user.ID, &user.Login, &user.Password)
	if err != nil {
		return user, err
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
		return user, err
	}

	return user, nil
}

func (r *userRepository) IsExistById(ctx context.Context, id int) bool {
	var exists bool
	query := "SELECT EXISTS (SELECT 1 FROM users WHERE id = $1)"
	err := r.Pool.QueryRow(ctx, query, id).Scan(&exists)
	if err != nil {
		return false
	}
	return exists
}
