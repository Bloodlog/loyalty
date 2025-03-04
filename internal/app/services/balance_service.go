package services

import (
	"context"
	"fmt"
	"gophermart/internal/app/apperrors"
	"gophermart/internal/app/dto"
	"gophermart/internal/app/entities"
	"gophermart/internal/app/repositories"
	"math"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type BalanceService interface {
	GetBalance(ctx context.Context, userID int) (dto.BalanceResponseBody, error)
	GetWithdrawals(ctx context.Context, userID int) ([]dto.WithdrawalsResponseBody, error)
	Withdraw(ctx context.Context, userID int, req dto.WithdrawBody) error
}

type balanceService struct {
	Pool               *pgxpool.Pool
	UserRepository     repositories.UserRepositoryInterface
	OrderRepository    repositories.OrderRepositoryInterface
	WithdrawRepository repositories.WithdrawRepositoryInterface
	roundingFactor     float64
}

func NewBalanceService(
	db *pgxpool.Pool,
	userRepository repositories.UserRepositoryInterface,
	orderRepository repositories.OrderRepositoryInterface,
	withdrawRepository repositories.WithdrawRepositoryInterface,
) BalanceService {
	const roundingFactor = 100
	return &balanceService{
		Pool:               db,
		UserRepository:     userRepository,
		OrderRepository:    orderRepository,
		WithdrawRepository: withdrawRepository,
		roundingFactor:     roundingFactor,
	}
}

func (o *balanceService) GetBalance(ctx context.Context, userID int) (dto.BalanceResponseBody, error) {
	var balanceResponse dto.BalanceResponseBody
	current, err := o.UserRepository.GetBalanceByUserID(ctx, nil, int64(userID))
	if err != nil {
		return balanceResponse, fmt.Errorf("failed GetBalance: %w", err)
	}
	withdrawn, err := o.WithdrawRepository.GetTotalWithdrawByUserID(ctx, userID)
	if err != nil {
		return balanceResponse, fmt.Errorf("failed GetTotalWithdrawByUserID: %w", err)
	}
	roundedAmount := math.Round(withdrawn*o.roundingFactor) / o.roundingFactor

	balanceResponse = dto.BalanceResponseBody{
		Current:   current,
		Withdrawn: roundedAmount,
	}

	return balanceResponse, nil
}

func (o *balanceService) GetWithdrawals(ctx context.Context, userID int) ([]dto.WithdrawalsResponseBody, error) {
	withdraws, err := o.WithdrawRepository.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed GetWithdrawals: %w", err)
	}
	response := make([]dto.WithdrawalsResponseBody, 0, len(withdraws))
	for _, withdraw := range withdraws {
		roundedAmount := math.Round(withdraw.Withdraw*o.roundingFactor) / o.roundingFactor
		response = append(response, dto.WithdrawalsResponseBody{
			Number:     strconv.Itoa(withdraw.OrderID),
			Withdrawaw: roundedAmount,
			CreatedAt:  withdraw.CreatedAt.Format(time.RFC3339),
		})
	}

	return response, nil
}

func (o *balanceService) Withdraw(ctx context.Context, userID int, req dto.WithdrawBody) error {
	tx, err := o.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer func(tx pgx.Tx, ctx context.Context) {
		_ = tx.Rollback(ctx)
	}(tx, ctx)

	current, err := o.UserRepository.GetBalanceByUserID(ctx, tx, int64(userID))
	if err != nil {
		return fmt.Errorf("failed GetBalanceByUserID: %w", err)
	}
	withdrawn, err := o.WithdrawRepository.GetTotalWithdrawByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed GetTotalWithdrawByUserID: %w", err)
	}
	if (current - withdrawn) < req.Sum {
		return apperrors.ErrBalanceNotEnought
	}
	number, err := strconv.ParseInt(req.OrderNumber, 10, 64)
	if err != nil {
		return fmt.Errorf("failed convert OrderNumber to int : %w", err)
	}

	roundedAmount := math.Round(req.Sum*o.roundingFactor) / o.roundingFactor
	withdrawOrder := entities.Withdraw{
		UserID:   int64(userID),
		OrderID:  int(number),
		Withdraw: roundedAmount,
	}

	err = o.WithdrawRepository.Save(ctx, tx, withdrawOrder)
	if err != nil {
		return fmt.Errorf("failed to save Withdraw: %w", err)
	}

	if withdrawOrder.Withdraw > 0 {
		var currentBalance float64
		currentBalance, err = o.UserRepository.GetBalanceByUserID(ctx, tx, withdrawOrder.UserID)
		if err != nil {
			return fmt.Errorf("failed to get current balance for user %d: %w", withdrawOrder.UserID, err)
		}
		newBalance := currentBalance - withdrawOrder.Withdraw
		err = o.UserRepository.UpdateBalanceByUserID(ctx, tx, newBalance, withdrawOrder.UserID)
		if err != nil {
			return fmt.Errorf(
				"failed to update user balance for user %d: %w",
				withdrawOrder.UserID,
				err,
			)
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
