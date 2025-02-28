package main

import (
	"context"
	"fmt"
	"gophermart/internal/app/command"
	"gophermart/internal/app/entities"
	"gophermart/internal/config"
	"gophermart/internal/logger"
	"gophermart/internal/server"
	"gophermart/internal/store"
	"log"

	"go.uber.org/zap"
)

func main() {
	loggerZap, err := logger.InitilazerLogger()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	if err = run(loggerZap); err != nil {
		loggerZap.Fatal("Application failed", zap.Error(err))
	}
}

func run(loggerZap *zap.SugaredLogger) error {
	cfg, err := config.ParseFlags()
	if err != nil {
		loggerZap.Info(err.Error(), "failed to parse flags")
		return fmt.Errorf("failed to parse flags: %w", err)
	}

	ctx := context.Background()
	storeDB, err := store.NewDB(ctx, cfg.DatabaseDsn)
	if err != nil {
		return fmt.Errorf("database error: %w", err)
	}

	sendQueue := make(chan *entities.Order, cfg.RateLimit)

	go func() {
		if err := command.ConfigureSendOrderHandler(storeDB.Pool, cfg, sendQueue, loggerZap); err != nil {
			log.Panicln(err)
		}
	}()
	if err := server.ConfigureServerHandler(storeDB.Pool, cfg, sendQueue, loggerZap); err != nil {
		log.Panicln(err)
	}
	return nil
}
