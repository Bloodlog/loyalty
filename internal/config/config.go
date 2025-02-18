package config

import "time"

type Config struct {
	DatabaseDsn        string
	HTTPAddress        string
	AccrualAddress     string
	AuthSecretKey      string
	AuthTokenExpired   time.Duration
	PollInterval       time.Duration
	RateLimit          int
	AgentTimeoutClient time.Duration
	AgentOrderLimit    int
	AgentDefaultRetry  time.Duration
}
