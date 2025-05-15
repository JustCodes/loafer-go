package loafergo

import "time"

const defaultRetryTimeout = 5 * time.Second

// Config defines settings shared across the manager and routes.
type Config struct {
	Logger       Logger
	RetryTimeout time.Duration
}

// loadConfig applies default values if not provided.
func loadConfig(cfg *Config) *Config {
	if cfg == nil {
		cfg = &Config{}
	}

	if cfg.RetryTimeout == 0 {
		cfg.RetryTimeout = defaultRetryTimeout
	}

	if cfg.Logger == nil {
		cfg.Logger = newDefaultLogger()
	}

	return cfg
}
