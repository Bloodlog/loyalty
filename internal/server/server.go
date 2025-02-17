package server

import (
	"fmt"
	"loyalty/internal/config"
	"loyalty/internal/routers"
	"net/http"

	"go.uber.org/zap"

	"github.com/jackc/pgx/v5/pgxpool"
)

func ConfigureServerHandler(db *pgxpool.Pool, cfg *config.Config, logger *zap.SugaredLogger) error {
	router := routers.ConfigureServerHandler(db, cfg)
	logger.Infoln("Start http server: ", cfg.HTTPAddress)
	err := http.ListenAndServe(cfg.HTTPAddress, router)
	if err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}
	return nil
}
