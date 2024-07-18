package loafergo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRouteWithMaxMessages(t *testing.T) {
	cfg := &RouteConfig{}
	var optConfigFns func(config *RouteConfig)
	optConfigFns = RouteWithMaxMessages(42)
	optConfigFns(cfg)
	assert.Equal(t, int32(42), cfg.maxMessages)
}

func TestRouteWithWaitTimeSeconds(t *testing.T) {
	cfg := &RouteConfig{}
	var optConfigFns func(config *RouteConfig)
	optConfigFns = RouteWithWaitTimeSeconds(42)
	optConfigFns(cfg)
	assert.Equal(t, int32(42), cfg.waitTimeSeconds)
}

func TestRouteWithVisibilityTimeout(t *testing.T) {
	cfg := &RouteConfig{}
	var optConfigFns func(config *RouteConfig)
	optConfigFns = RouteWithVisibilityTimeout(42)
	optConfigFns(cfg)
	assert.Equal(t, int32(42), cfg.visibilityTimeout)
}
