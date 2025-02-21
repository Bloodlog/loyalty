package services

import (
	"context"
	"errors"
	"fmt"
	"loyalty/internal/app/apperrors"
	"loyalty/internal/app/dto"
	"loyalty/internal/app/entities"
	"loyalty/internal/app/repositories"
	"strconv"
	"time"
)

type OrderService interface {
	GetOrdersByUserID(ctx context.Context, userID int) ([]dto.OrdersResponseBody, error)
	SaveOrder(ctx context.Context, req dto.OrderBody) error
}

type orderService struct {
	OrderRepository repositories.OrderRepositoryInterface
}

func NewOrderService(
	orderRepository repositories.OrderRepositoryInterface,
) OrderService {
	return &orderService{
		OrderRepository: orderRepository,
	}
}

func (o *orderService) GetOrdersByUserID(ctx context.Context, userID int) ([]dto.OrdersResponseBody, error) {
	orders, err := o.OrderRepository.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get order by user id: %w", err)
	}
	response := make([]dto.OrdersResponseBody, 0, len(orders))
	for i := range orders {
		order := &orders[i]
		status := entities.GetStatusName(int(order.StatusID))
		var accrual *float64
		if order.Accrual.Valid {
			accrual = &order.Accrual.Float64
		}
		response = append(response, dto.OrdersResponseBody{
			Number:     strconv.Itoa(order.OrderID),
			Status:     status,
			Accrual:    accrual,
			UploadedAt: order.UpdatedAt.Format(time.RFC3339),
		})
	}

	return response, nil
}

func (o *orderService) SaveOrder(ctx context.Context, req dto.OrderBody) error {
	if err := o.validateOrder(ctx, req.OrderNumber, req.UserID); err != nil {
		return err
	}

	order := entities.Order{
		UserID:   req.UserID,
		StatusID: int16(req.StatusID),
		OrderID:  int(req.OrderNumber),
	}
	err := o.OrderRepository.Store(ctx, &order)
	if err != nil {
		return fmt.Errorf("failed to SaveOrder: %w", err)
	}

	return nil
}

func (o *orderService) validateOrder(ctx context.Context, orderNumber int64, userID int64) error {
	order, err := o.OrderRepository.GetByOrderNumber(ctx, orderNumber)
	if err == nil {
		if order.UserID != userID {
			return apperrors.ErrDuplicateOrderIDAnotherUserID
		}
		return apperrors.ErrDuplicateOrderID
	}

	if errors.Is(err, apperrors.ErrOrderNotFound) {
		return nil
	}

	return fmt.Errorf("failed to validate order: %w", err)
}
