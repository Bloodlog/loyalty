package services

import (
	"context"
	"database/sql"
	"fmt"
	"loyalty/internal/app/entities"
	"loyalty/internal/app/services/accrual"
	"time"

	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"

	"loyalty/internal/app/repositories"
	"loyalty/internal/config"
)

type AccrualService interface {
	SendOrder(ctx context.Context, order *entities.Order) error
	GetFreshOrders(ctx context.Context, limit int) ([]entities.Order, error)
}

type accrualService struct {
	OrderRepository repositories.OrderRepositoryInterface
	Client          *resty.Client
	Cfg             *config.Config
	Logger          *zap.SugaredLogger
}

func NewAccrualService(
	orderRepository repositories.OrderRepositoryInterface,
	client *resty.Client,
	cfg *config.Config,
	logger *zap.SugaredLogger,
) AccrualService {
	return &accrualService{
		OrderRepository: orderRepository,
		Client:          client,
		Cfg:             cfg,
		Logger:          logger,
	}
}

func (u *accrualService) GetFreshOrders(ctx context.Context, limit int) ([]entities.Order, error) {
	orders, err := u.OrderRepository.GetFreshOrders(ctx, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to GetFreshOrders: %w", err)
	}
	return orders, nil
}

func (u *accrualService) SendOrder(ctx context.Context, order *entities.Order) error {
	orderID := int64(order.OrderID)
	order.UpdatedAt = time.Now()
	orderResponse, err := accrual.SendOrder(u.Client, orderID)
	if err != nil {
		u.Logger.Infoln(err)
		return fmt.Errorf("failed to SendOrder to accrual: %w", err)
	}

	statusID, _ := entities.GetStatusIDByName(orderResponse.Status)
	if orderResponse.Accrual != nil {
		order.Accrual = sql.NullFloat64{Float64: *orderResponse.Accrual, Valid: true}
	} else {
		order.Accrual = sql.NullFloat64{Valid: false}
	}
	order.StatusID = int16(statusID)
	err = u.OrderRepository.UpdateOrder(ctx, order)
	if err != nil {
		u.Logger.Infoln(err)
		return fmt.Errorf("failed to UpdateOrder with status: %w", err)
	}

	return nil
}
