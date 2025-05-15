package loafergo_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	loafergo "github.com/justcodes/loafer-go/v2"
	"github.com/justcodes/loafer-go/v2/fake"
)

func BenchmarkManager_Run(b *testing.B) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	message := new(fake.Message)
	router := new(fake.Router)
	router.On("Configure", mock.Anything).Return(nil)
	router.On("WorkerPoolSize", mock.Anything).Return(int32(4))
	router.On("RunMode", mock.Anything).Return(loafergo.Parallel)
	router.On("GetMessages", mock.Anything).Return([]loafergo.Message{message}, nil).Maybe()
	router.On("HandlerMessage", mock.Anything, message).Return(nil)
	router.On("Commit", mock.Anything, message).Return(nil)

	logger := new(fake.Logger)
	logger.On("Log", mock.Anything).Return()

	manager := loafergo.NewManager(&loafergo.Config{
		Logger:       logger,
		RetryTimeout: 10 * time.Millisecond,
	})
	manager.RegisterRoute(router)

	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	b.ResetTimer()
	err := manager.Run(ctx)
	b.StopTimer()

	assert.NoError(b, err)
}
