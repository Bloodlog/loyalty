package services

import (
	"context"
	"fmt"
	"loyalty/internal/app/apperrors"
	"loyalty/internal/app/dto"
	"loyalty/internal/app/entities"
	"loyalty/internal/app/repositories"
	"math"
	"strconv"
	"time"
)

type BalanceService interface {
	GetBalance(ctx context.Context, userID int) (dto.BalanceResponseBody, error)
	GetWithdrawals(ctx context.Context, userID int) ([]dto.WithdrawalsResponseBody, error)
	Withdraw(ctx context.Context, userID int, req dto.WithdrawBody) error
}

type balanceService struct {
	UserRepository     repositories.UserRepositoryInterface
	OrderRepository    repositories.OrderRepositoryInterface
	WithdrawRepository repositories.WithdrawRepositoryInterface
}

func NewBalanceService(
	userRepository repositories.UserRepositoryInterface,
	orderRepository repositories.OrderRepositoryInterface,
	withdrawRepository repositories.WithdrawRepositoryInterface,
) BalanceService {
	return &balanceService{
		UserRepository:     userRepository,
		OrderRepository:    orderRepository,
		WithdrawRepository: withdrawRepository,
	}
}

func (o *balanceService) GetBalance(ctx context.Context, userID int) (dto.BalanceResponseBody, error) {
	var balanceResponse dto.BalanceResponseBody
	current, err := o.UserRepository.GetBalanceByUserID(ctx, userID)
	if err != nil {
		return balanceResponse, fmt.Errorf("failed GetBalance: %w", err)
	}
	withdrawn, err := o.WithdrawRepository.GetTotalWithdrawByUserID(ctx, userID)
	if err != nil {
		return balanceResponse, fmt.Errorf("failed GetTotalWithdrawByUserID: %w", err)
	}
	balanceResponse = dto.BalanceResponseBody{
		Current:   current,
		Withdrawn: int64(withdrawn),
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
		response = append(response, dto.WithdrawalsResponseBody{
			Number:     strconv.Itoa(withdraw.OrderID),
			Withdrawaw: withdraw.Withdraw,
			CreatedAt:  withdraw.CreatedAt.Format(time.RFC3339),
		})
	}

	return response, nil
}

func (o *balanceService) Withdraw(ctx context.Context, userID int, req dto.WithdrawBody) error {
	current, err := o.UserRepository.GetBalanceByUserID(ctx, userID)
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

	roundedAmount := math.Round(req.Sum*100) / 100
	withdrawOrder := entities.Withdraw{
		UserID:   int64(userID),
		OrderID:  int(number),
		Withdraw: roundedAmount,
	}
	err = o.WithdrawRepository.StoreAndUpdateBalance(ctx, withdrawOrder)
	if err != nil {
		return fmt.Errorf("failed to save Withdraw: %w", err)
	}

	return nil
}
