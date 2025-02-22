package database

import (
	"context"
	"fmt"
	"gophermart/internal/config"
	"gophermart/internal/db/migrations"

	"go.uber.org/zap"

	"github.com/jackc/pgx/v5/pgxpool"
)

func ConfigureDatabase(
	ctx context.Context,
	cfg *config.Config,
	logger *zap.SugaredLogger,
) (*pgxpool.Pool, error) {
	poolCfg, err := pgxpool.ParseConfig(cfg.DatabaseDsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse the DSN: %w", err)
	}
	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize a connection pool: %w", err)
	}

	err = migrations.Migrate(ctx, pool, logger)
	if err != nil {
		return nil, fmt.Errorf("error migrate: %w", err)
	}

	return pool, nil
}
