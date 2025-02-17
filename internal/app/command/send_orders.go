package command

import (
	"loyalty/internal/app/handlers"
	"loyalty/internal/app/repositories"
	"loyalty/internal/app/services"
	"loyalty/internal/app/services/send_orders"
	"loyalty/internal/config"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

func ConfigureSendOrderHandler(db *pgxpool.Pool, cfg *config.Config, logger *zap.SugaredLogger) error {
	client := send_orders.NewClient(cfg.AccrualAddress, cfg.AgentTimeoutClient)
	orderRepository := repositories.NewOrderRepository(db)
	sendOrdersService := services.NewSendOrdersService(orderRepository, client, cfg, logger)
	sendOrderHandler := handlers.NewSendOrderHandler(sendOrdersService, cfg, logger)
	logger.Infoln("Start accrual agent interval:", cfg.PollInterval)
	err := sendOrderHandler.SendUserOrders()
	if err != nil {
		return err
	}

	return nil
}
