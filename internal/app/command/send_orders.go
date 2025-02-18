package command

import (
	"fmt"
	"loyalty/internal/app/handlers"
	"loyalty/internal/app/repositories"
	"loyalty/internal/app/services"
	"loyalty/internal/app/services/accrual"
	"loyalty/internal/config"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

func ConfigureSendOrderHandler(db *pgxpool.Pool, cfg *config.Config, logger *zap.SugaredLogger) error {
	client := accrual.NewClient(cfg.AccrualAddress, cfg.AgentTimeoutClient)
	orderRepository := repositories.NewOrderRepository(db)
	sendOrdersService := services.NewAccrualService(orderRepository, client, cfg, logger)
	sendOrderHandler := handlers.NewSendOrderHandler(sendOrdersService, cfg, logger)
	logger.Infoln("Start accrual agent interval:", cfg.PollInterval)
	err := sendOrderHandler.SendUserOrders()
	if err != nil {
		return fmt.Errorf("failed send user orders: %w", err)
	}

	return nil
}
