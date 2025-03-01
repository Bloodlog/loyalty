package services

import (
	"context"
	"errors"
	"fmt"
	"gophermart/internal/app/apperrors"
	"gophermart/internal/app/dto"
	"gophermart/internal/app/entities"
	"gophermart/internal/app/repositories"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OrderService interface {
	GetOrdersByUserID(ctx context.Context, userID int) ([]dto.OrdersResponseBody, error)
	SaveOrder(ctx context.Context, req dto.OrderBody) error
}

type orderService struct {
	Pool            *pgxpool.Pool
	OrderRepository repositories.OrderRepositoryInterface
	JobRepository   repositories.JobRepositoryInterface
}

func NewOrderService(
	db *pgxpool.Pool,
	orderRepository repositories.OrderRepositoryInterface,
	jobRepository repositories.JobRepositoryInterface,
) OrderService {
	return &orderService{
		Pool:            db,
		OrderRepository: orderRepository,
		JobRepository:   jobRepository,
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
		return fmt.Errorf("validate failed: %w", err)
	}

	order := entities.Order{
		UserID:   req.UserID,
		StatusID: int16(req.StatusID),
		OrderID:  int(req.OrderNumber),
	}

	tx, err := o.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to SaveOrder: %w", err)
	}
	defer func(tx pgx.Tx, ctx context.Context) {
		err = tx.Rollback(ctx)
	}(tx, ctx)

	ID, err := o.OrderRepository.Store(ctx, tx, &order)
	if err != nil {
		return fmt.Errorf("failed to SaveOrder: %w", err)
	}
	job := entities.Job{
		OrderID: int64(ID),
	}
	err = o.JobRepository.SaveJob(ctx, tx, &job)
	if err != nil {
		return fmt.Errorf("failed to Save job: %w", err)
	}
	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to SaveOrder: %w", err)
	}

	return nil
}

func (o *orderService) validateOrder(ctx context.Context, orderNumber int64, userID int64) error {
	order, err := o.OrderRepository.GetByOrderNumber(ctx, orderNumber)
	if err != nil {
		if errors.Is(err, apperrors.ErrOrderNotFound) {
			return nil
		}

		return fmt.Errorf("failed to validate order: %w", err)
	}
	if order.UserID != userID {
		return apperrors.ErrDuplicateOrderIDAnotherUserID
	}
	return apperrors.ErrDuplicateOrderID
}
