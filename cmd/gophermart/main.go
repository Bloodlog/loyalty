package main

import (
	"context"
	"fmt"
	"log"
	"loyalty/internal/app/command"
	"loyalty/internal/app/entities"
	"loyalty/internal/config"
	database "loyalty/internal/db"
	"loyalty/internal/logger"
	"loyalty/internal/server"

	"go.uber.org/zap"
)

func main() {
	loggerZap, err := logger.InitilazerLogger()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	if err := run(loggerZap); err != nil {
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
	db, err := database.ConfigureDatabase(ctx, cfg, loggerZap)
	if err != nil {
		return fmt.Errorf("database error: %w", err)
	}

	sendQueue := make(chan *entities.Order, cfg.RateLimit)

	go func() {
		if err := command.ConfigureSendOrderHandler(db, cfg, sendQueue, loggerZap); err != nil {
			log.Panicln(err)
		}
	}()
	if err := server.ConfigureServerHandler(db, cfg, sendQueue, loggerZap); err != nil {
		log.Panicln(err)
	}
	return nil
}
