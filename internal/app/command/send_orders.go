package command

import (
	"fmt"
	"gophermart/internal/app/handlers"
	"gophermart/internal/app/repositories"
	"gophermart/internal/app/services"
	"gophermart/internal/app/services/accrual"
	"gophermart/internal/config"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

func ConfigureSendOrderHandler(
	db *pgxpool.Pool,
	cfg *config.Config,
	logger *zap.SugaredLogger,
) error {
	client := accrual.NewClient(cfg.AccrualAddress, cfg.AgentTimeoutClient)
	orderRepository := repositories.NewOrderRepository(db)
	userRepository := repositories.NewUserRepository(db)
	jobRepository := repositories.NewJobRepository(db)
	sendOrdersService := services.NewAccrualService(
		db,
		jobRepository,
		orderRepository,
		userRepository,
		client,
		cfg,
		logger,
	)
	sendOrderHandler := handlers.NewSendOrderHandler(sendOrdersService, cfg, logger)
	logger.Infoln("Start accrual agent interval:", cfg.PollInterval)
	err := sendOrderHandler.SendUserOrders()
	if err != nil {
		return fmt.Errorf("failed send user orders: %w", err)
	}

	return nil
}
