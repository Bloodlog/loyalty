package config

import (
	"flag"
	"fmt"
	"os"
	"time"
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

	agentOrderLimit := 5
	agentTimeoutClient := 10 * time.Second
	if agentOrderLimit > int(agentTimeoutClient/time.Second) {
		return nil, fmt.Errorf(
			"AgentOrderLimit (%d) превышает допустимое количество для заданного AgentTimeoutClient (%s)",
			agentOrderLimit,
			agentTimeoutClient,
		)
	}

	return &Config{
		DatabaseDsn:        databaseDsn,
		HTTPAddress:        httpAddress,
		AccrualAddress:     accrualAddress,
		AuthSecretKey:      applicationKey,
		AuthTokenExpired:   time.Hour * 3,
		PollInterval:       time.Minute * 1,
		RateLimit:          1,
		AgentTimeoutClient: agentTimeoutClient,
		AgentOrderLimit:    agentOrderLimit,
	}, nil
}

func validateUnknownArgs(unknownArgs []string) error {
	if len(unknownArgs) > 0 {
		return fmt.Errorf("unknown flags or arguments detected: %v", unknownArgs)
	}
	return nil
}

func getStringValue(env, flag string) string {
	if envValue, exists := os.LookupEnv(env); exists {
		return envValue
	} else {
		return flag
	}
}
