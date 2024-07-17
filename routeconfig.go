package loafergo

const (
	defaultExtensionLimit    = 2
	defaultVisibilityTimeout = int32(30)
	defaultMaxMessages       = int32(10)
	defaultWaitTimeSeconds   = int32(10)
)

// RouteConfig are discrete set of route options that are valid for loading the route configuration
type RouteConfig struct {
	visibilityTimeout int32
	maxMessages       int32
	extensionLimit    int
	waitTimeSeconds   int32
}

func loadDefaultRouteConfig() *RouteConfig {
	return &RouteConfig{
		visibilityTimeout: defaultVisibilityTimeout,
		maxMessages:       defaultMaxMessages,
		extensionLimit:    defaultExtensionLimit,
		waitTimeSeconds:   defaultWaitTimeSeconds,
	}
}

// LoadRouteConfigFunc is a type alias for RouteConfig functional config
type LoadRouteConfigFunc func(config *RouteConfig)

// RouteWithVisibilityTimeout is a helper function to construct functional options that sets visibility Timeout value
// on config's Route. If multiple RouteWithVisibilityTimeout calls are made,
// the last call overrides the previous call values.
func RouteWithVisibilityTimeout(v int32) LoadRouteConfigFunc {
	return func(rc *RouteConfig) {
		rc.visibilityTimeout = v
	}
}

// RouteWithMaxMessages is a helper function to construct functional options that sets Max Messages value
// on config's Route. If multiple RouteWithMaxMessages calls are made,
// the last call overrides the previous call values.
func RouteWithMaxMessages(v int32) LoadRouteConfigFunc {
	return func(rc *RouteConfig) {
		rc.maxMessages = v
	}
}

// RouteWithWaitTimeSeconds is a helper function to construct functional options that sets Wait Time Seconds value
// on config's Route. If multiple RouteWithWaitTimeSeconds calls are made,
// the last call overrides the previous call values.
func RouteWithWaitTimeSeconds(v int32) LoadRouteConfigFunc {
	return func(rc *RouteConfig) {
		rc.waitTimeSeconds = v
	}
}
