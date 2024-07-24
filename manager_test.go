package loafergo_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	loafergo "github.com/justcodes/loafer-go"
	"github.com/justcodes/loafer-go/fake"
)

func createTestConfig() *loafergo.Config {
	return &loafergo.Config{
		Logger: loafergo.LoggerFunc(func(args ...interface{}) {}),
	}
}

func TestManager_RegisterRoute(t *testing.T) {
	config := createTestConfig()

	manager := loafergo.NewManager(config)
	mockRoute := &fake.Router{}

	manager.RegisterRoute(mockRoute)
	assert.Len(t, manager.GetRoutes(), 1)
}

func TestManager_RegisterRoutes(t *testing.T) {
	config := createTestConfig()
	manager := loafergo.NewManager(config)
	mockRoutes := []loafergo.Router{
		&fake.Router{},
		&fake.Router{},
	}

	manager.RegisterRoutes(mockRoutes)
	assert.Len(t, manager.GetRoutes(), 2)
}

func TestManager_Run(t *testing.T) {
	t.Run("Should return error when configure error", func(t *testing.T) {
		ctx := context.Background()
		config := createTestConfig()
		manager := loafergo.NewManager(config)
		mockRoute := &fake.Router{}

		mockRoute.On("Configure", context.Background()).
			Return(fmt.Errorf("error")).
			Once()

		manager.RegisterRoute(mockRoute)
		err := manager.Run(ctx)
		assert.NotNil(t, err)
	})

	t.Run("Should return error no routes", func(t *testing.T) {
		ctx := context.Background()
		config := createTestConfig()
		manager := loafergo.NewManager(config)

		err := manager.Run(ctx)
		assert.NotNil(t, err)
		assert.ErrorIs(t, err, loafergo.ErrNoRoute)
	})
}
