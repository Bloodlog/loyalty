package command

import (
	"fmt"
	"gophermart/internal/app/entities"
	"gophermart/internal/app/handlers"
	"gophermart/internal/app/repositories"
	"gophermart/internal/app/services"
	"gophermart/internal/app/services/accrual"
	"gophermart/internal/config"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

func ConfigureSendOrderHandler(db *pgxpool.Pool, cfg *config.Config, queue chan *entities.Order, logger *zap.SugaredLogger) error {
	client := accrual.NewClient(cfg.AccrualAddress, cfg.AgentTimeoutClient)
	orderRepository := repositories.NewOrderRepository(db)
	sendOrdersService := services.NewAccrualService(orderRepository, client, cfg, logger)
	sendOrderHandler := handlers.NewSendOrderHandler(sendOrdersService, cfg, queue, logger)
	logger.Infoln("Start accrual agent interval:", cfg.PollInterval)
	err := sendOrderHandler.SendUserOrders()
	if err != nil {
		return fmt.Errorf("failed send user orders: %w", err)
	}

	return nil
}
