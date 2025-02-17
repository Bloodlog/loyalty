package services

import (
	"context"
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
	var response []dto.OrdersResponseBody
	orders, err := o.OrderRepository.GetByUserID(ctx, userID)
	if err != nil {
		return response, err
	}
	for _, order := range orders {
		status := entities.GetStatusName(int(order.StatusId))
		var accrual *float64
		if order.Accrual.Valid {
			accrual = &order.Accrual.Float64
		}
		response = append(response, dto.OrdersResponseBody{
			Number:     strconv.Itoa(order.OrderId),
			Status:     status,
			Accrual:    accrual,
			UploadedAt: order.UpdatedAt.Format(time.RFC3339),
		})
	}

	return response, nil
}

func (o *orderService) SaveOrder(ctx context.Context, req dto.OrderBody) error {
	order := entities.Order{
		UserID:   req.UserId,
		StatusId: int16(req.StatusId),
		OrderId:  int(req.OrderNumber),
	}
	err := o.OrderRepository.Store(ctx, order)
	if err != nil {
		return err
	}

	return nil
}
