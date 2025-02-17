package logger

import (
	"fmt"

	"go.uber.org/zap"
)

func InitilazerLogger() (*zap.SugaredLogger, error) {
	configLogger := zap.NewDevelopmentConfig()
	configLogger.Level = zap.NewAtomicLevelAt(zap.InfoLevel)

	logger, err := configLogger.Build()
	if err != nil {
		return nil, fmt.Errorf("logger initialization failed: %w", err)
	}

	return logger.Sugar(), err
}
