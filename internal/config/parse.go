package config

import (
	"flag"
	"fmt"
	"os"
	"time"
)

const (
	defaultAgentTimeout        = 10 * time.Second
	defaultAuthTokenExpiration = 3 * time.Hour
	defaultPollInterval        = time.Minute
	defaultAgentOrderLimit     = 5
	defaultAgentRetry          = 30 * time.Second
)

func ParseFlags() (*Config, error) {
	httpAddressFlag := flag.String("a", "", "адрес и порт запуска сервиса")
	databaseDsnFlag := flag.String("d", "", "адрес подключения к базе данных")
	accrualAddressFlag := flag.String("r", "", "адрес системы расчёта начислений")

	flag.Parse()

	uknownArguments := flag.Args()
	if err := validateUnknownArgs(uknownArguments); err != nil {
		return nil, fmt.Errorf("read flag UnknownArgs: %w", err)
	}
	httpAddress := getStringValue("RUN_ADDRESS", *httpAddressFlag)
	databaseDsn := getStringValue("DATABASE_URI", *databaseDsnFlag)
	accrualAddress := getStringValue("ACCRUAL_SYSTEM_ADDRESS", *accrualAddressFlag)
	applicationKey := "supersecretkey"

	agentOrderLimit := defaultAgentOrderLimit
	agentTimeoutClient := defaultAgentTimeout
	if agentOrderLimit > int(agentTimeoutClient/time.Second) {
		return nil, fmt.Errorf(
			"AgentOrderLimit (%d) превышает допустимое количество для заданного AgentTimeoutClient (%s)",
			agentOrderLimit,
			agentTimeoutClient,
		)
	}
	rateLimit := 1

	return &Config{
		DatabaseDsn:        databaseDsn,
		HTTPAddress:        httpAddress,
		AccrualAddress:     accrualAddress,
		AuthSecretKey:      applicationKey,
		AuthTokenExpired:   defaultAuthTokenExpiration,
		PollInterval:       defaultPollInterval,
		RateLimit:          rateLimit,
		AgentTimeoutClient: agentTimeoutClient,
		AgentOrderLimit:    agentOrderLimit,
		AgentDefaultRetry:  defaultAgentRetry,
	}, nil
}

func validateUnknownArgs(unknownArgs []string) error {
	if len(unknownArgs) > 0 {
		return fmt.Errorf("unknown flags or arguments detected: %v", unknownArgs)
	}
	return nil
}

func getStringValue(env, flagValue string) string {
	if envValue, exists := os.LookupEnv(env); exists {
		return envValue
	} else {
		return flagValue
	}
}
