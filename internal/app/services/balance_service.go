package services

import (
	"context"
	"loyalty/internal/app/apperrors"
	"loyalty/internal/app/dto"
	"loyalty/internal/app/entities"
	"loyalty/internal/app/repositories"
	"strconv"
	"time"
)

type BalanceService interface {
	GetBalance(ctx context.Context, userId int) (dto.BalanceResponseBody, error)
	GetWithdrawals(ctx context.Context, userId int) ([]dto.WithdrawalsResponseBody, error)
	Withdraw(ctx context.Context, userId int, req dto.WithdrawBody) error
}

type balanceService struct {
	OrderRepository    repositories.OrderRepositoryInterface
	WithdrawRepository repositories.WithdrawRepositoryInterface
}

func NewBalanceService(
	orderRepository repositories.OrderRepositoryInterface,
	withdrawRepository repositories.WithdrawRepositoryInterface,
) BalanceService {
	return &balanceService{
		OrderRepository:    orderRepository,
		WithdrawRepository: withdrawRepository,
	}
}

func (o *balanceService) GetBalance(ctx context.Context, userId int) (dto.BalanceResponseBody, error) {
	var balanceResponse dto.BalanceResponseBody
	current, err := o.OrderRepository.GetTotalAccrualByUserID(ctx, userId)
	if err != nil {
		return balanceResponse, err
	}
	withdrawn, err := o.WithdrawRepository.GetTotalWithdrawByUserID(ctx, userId)
	if err != nil {
		return balanceResponse, err
	}
	balanceResponse = dto.BalanceResponseBody{
		Current:   current,
		Withdrawn: int64(withdrawn),
	}

	return balanceResponse, nil
}

func (o *balanceService) GetWithdrawals(ctx context.Context, userId int) ([]dto.WithdrawalsResponseBody, error) {
	var response []dto.WithdrawalsResponseBody
	withdraws, err := o.WithdrawRepository.GetByUserID(ctx, userId)
	if err != nil {
		return response, err
	}

	for _, withdraw := range withdraws {
		response = append(response, dto.WithdrawalsResponseBody{
			Number:     strconv.Itoa(withdraw.OrderId),
			Withdrawaw: withdraw.Withdraw,
			CreatedAt:  withdraw.CreatedAt.Format(time.RFC3339),
		})
	}

	return response, nil
}

func (o *balanceService) Withdraw(ctx context.Context, userId int, req dto.WithdrawBody) error {
	current, err := o.OrderRepository.GetTotalAccrualByUserID(ctx, userId)
	if err != nil {
		return err
	}
	withdrawn, err := o.WithdrawRepository.GetTotalWithdrawByUserID(ctx, userId)
	if err != nil {
		return err
	}
	if (current - withdrawn) < float64(req.Sum) {
		return apperrors.ErrBalanceNotEnought
	}

	withdrawOrder := entities.Withdraw{
		UserID:   int64(userId),
		OrderId:  int(req.OrderNumber),
		Withdraw: float64(req.Sum),
	}
	err = o.WithdrawRepository.Store(ctx, withdrawOrder)
	if err != nil {
		return err
	}

	return nil
}
