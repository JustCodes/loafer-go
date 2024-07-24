package loafergo

import "time"

const defaultRetryTimeout = 5 * time.Second

// Config defines the loafer Manager configuration
type Config struct {
	Logger Logger

	// RetryTimeout is used when the Route GetMessages method returns error
	// By default the retry timeout is 5s
	RetryTimeout time.Duration
}
