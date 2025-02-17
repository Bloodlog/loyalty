package services

import (
	"context"
	"database/sql"
	"errors"
	"loyalty/internal/app/entities"
	"loyalty/internal/app/services/send_orders"
	"time"

	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"

	"loyalty/internal/app/repositories"
	"loyalty/internal/config"
)

type SendOrdersService interface {
	SendOrder(ctx context.Context, order entities.Order) error
	GetFreshOrders(ctx context.Context, limit int) ([]entities.Order, error)
}

type sendOrdersService struct {
	OrderRepository repositories.OrderRepositoryInterface
	Client          *resty.Client
	Cfg             *config.Config
	Logger          *zap.SugaredLogger
}

func NewSendOrdersService(
	orderRepository repositories.OrderRepositoryInterface,
	client *resty.Client,
	cfg *config.Config,
	logger *zap.SugaredLogger,
) SendOrdersService {
	return &sendOrdersService{
		OrderRepository: orderRepository,
		Client:          client,
		Cfg:             cfg,
		Logger:          logger,
	}
}

func (u *sendOrdersService) GetFreshOrders(ctx context.Context, limit int) ([]entities.Order, error) {
	orders, err := u.OrderRepository.GetFreshOrders(ctx, limit)
	if err != nil {
		return nil, err
	}
	return orders, nil
}

func (u *sendOrdersService) SendOrder(ctx context.Context, order entities.Order) error {
	orderId := int64(order.OrderId)
	order.UpdatedAt = time.Now()
	orderResponse, retryAfter, err := send_orders.SendOrder(u.Client, orderId)
	if err != nil {
		u.Logger.Infoln(err)
		if errors.Is(err, send_orders.ErrTooManyRequests) {
			var dur time.Duration
			dur = 30 * time.Second
			if retryAfter != 0 {
				dur = time.Duration(retryAfter) * time.Second
			}
			order.NextAttempt = sql.NullTime{Time: time.Now().Add(dur), Valid: true}
			order.Attempts = order.Attempts + 1
			err = u.OrderRepository.UpdateOrder(ctx, order)
			if err != nil {
				return err
			}
		}
		return err
	}

	statusId, _ := entities.GetStatusIdByName(orderResponse.Status)
	if orderResponse.Accrual != nil {
		order.Accrual = sql.NullFloat64{Float64: *orderResponse.Accrual, Valid: true}
	} else {
		order.Accrual = sql.NullFloat64{Valid: false}
	}
	order.StatusId = int16(statusId)
	err = u.OrderRepository.UpdateOrder(ctx, order)
	if err != nil {
		u.Logger.Infoln(err)
		return err
	}

	return nil
}
