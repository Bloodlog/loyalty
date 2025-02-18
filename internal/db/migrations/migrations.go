package migrations

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Migrate(ctx context.Context, conn *pgxpool.Pool, logger *zap.SugaredLogger) error {
	tx, err := conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("failed to start a transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil {
			if !errors.Is(err, sql.ErrTxDone) {
				logger.Infoln("failed to rollback the transaction: %v", err)
			}
		}
	}()
	createSchemaStmts := []string{
		`	CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
			login VARCHAR(200) NOT NULL UNIQUE,
			password VARCHAR(200) NOT NULL
			)`,

		`CREATE TABLE IF NOT EXISTS orders (
			id BIGINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
			order_number BIGINT NOT NULL UNIQUE,
			status_id INT NOT NULL,
			user_id INT NOT NULL,
			accrual FLOAT NULL,
			next_attempt TIMESTAMP NULL,
			attempts INTEGER DEFAULT 0,
			created_at TIMESTAMP DEFAULT now(),
			updated_at TIMESTAMP DEFAULT now(),
			CONSTRAINT fk_orders_user FOREIGN KEY (user_id) REFERENCES users(id)
			)`,
		`CREATE TABLE IF NOT EXISTS withdraws (
			id BIGINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
			order_number BIGINT NOT NULL UNIQUE,
			user_id INT NOT NULL,
			withdraw FLOAT NOT NULL,
			created_at TIMESTAMP DEFAULT now(),
			CONSTRAINT fk_orders_user FOREIGN KEY (user_id) REFERENCES users(id)
			)`,
	}

	for _, stmt := range createSchemaStmts {
		if _, err := tx.Exec(ctx, stmt); err != nil {
			return fmt.Errorf("failed to execute statement `%s`: %w", stmt, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
