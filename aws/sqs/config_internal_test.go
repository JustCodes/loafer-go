package sqs

import (
	"testing"

	"github.com/stretchr/testify/assert"

	loafergo "github.com/justcodes/loafer-go/v2"
)

func TestRouteWithMaxMessages(t *testing.T) {
	cfg := &RouteConfig{}
	var optConfigFns func(config *RouteConfig)
	optConfigFns = RouteWithMaxMessages(42)
	optConfigFns(cfg)
	assert.Equal(t, int32(42), cfg.maxMessages)
}

func TestRouteWithWaitTimeSeconds(t *testing.T) {
	cfg := loadDefaultRouteConfig()
	var optConfigFns func(config *RouteConfig)
	optConfigFns = RouteWithWaitTimeSeconds(42)
	optConfigFns(cfg)
	assert.Equal(t, int32(42), cfg.waitTimeSeconds)
}

func TestRouteWithVisibilityTimeout(t *testing.T) {
	t.Run("With custom visibility timeout > defaultVisibilityTimeoutControl", func(t *testing.T) {
		cfg := loadDefaultRouteConfig()
		var optConfigFns func(config *RouteConfig)
		optConfigFns = RouteWithVisibilityTimeout(42)
		optConfigFns(cfg)
		assert.Equal(t, int32(42), cfg.visibilityTimeout)
	})

	t.Run("With custom visibility timeout <= defaultVisibilityTimeoutControl", func(t *testing.T) {
		cfg := loadDefaultRouteConfig()
		var optConfigFns func(config *RouteConfig)
		optConfigFns = RouteWithVisibilityTimeout(defaultVisibilityTimeoutControl)
		optConfigFns(cfg)
		assert.Equal(t, int32(defaultVisibilityTimeoutControl+1), cfg.visibilityTimeout)
	})

}

func TestRouteWithWorkerPoolSize(t *testing.T) {
	cfg := loadDefaultRouteConfig()
	var optConfigFns func(config *RouteConfig)
	optConfigFns = RouteWithWorkerPoolSize(42)
	optConfigFns(cfg)
	assert.Equal(t, int32(42), cfg.workerPoolSize)
}

func TestRouteWithRunMode(t *testing.T) {
	t.Run("Set to PerGroupID", func(t *testing.T) {
		cfg := &RouteConfig{}
		var optConfigFns func(config *RouteConfig)
		optConfigFns = RouteWithRunMode(loafergo.PerGroupID)
		optConfigFns(cfg)
		assert.Equal(t, loafergo.PerGroupID, cfg.runMode)
	})

	t.Run("Set to Parallel", func(t *testing.T) {
		cfg := &RouteConfig{}
		var optConfigFns func(config *RouteConfig)
		optConfigFns = RouteWithRunMode(loafergo.Parallel)
		optConfigFns(cfg)
		assert.Equal(t, loafergo.Parallel, cfg.runMode)
	})
}

func TestRouteWithCustomGroupFields(t *testing.T) {
	t.Run("Set custom fields", func(t *testing.T) {
		cfg := &RouteConfig{}
		fields := []string{"seller_id", "marketplace_id"}
		var optConfigFns func(config *RouteConfig)
		optConfigFns = RouteWithCustomGroupFields(fields)
		optConfigFns(cfg)
		assert.Equal(t, fields, cfg.customGroupFields)
	})

	t.Run("Set empty field list", func(t *testing.T) {
		cfg := &RouteConfig{}
		var fields []string
		var optConfigFns func(config *RouteConfig)
		optConfigFns = RouteWithCustomGroupFields(fields)
		optConfigFns(cfg)
		assert.Empty(t, cfg.customGroupFields)
	})
}
