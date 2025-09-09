package loafergo_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	loafergo "github.com/justcodes/loafer-go/v2"
	"github.com/justcodes/loafer-go/v2/fake"
)

func TestManager_Run_NoRoutes(t *testing.T) {
	m := loafergo.NewManager(nil)
	err := m.Run(context.Background())
	assert.Equal(t, loafergo.ErrNoRoute, err)
}

func TestManager_RegisterRoute_And_GetRoutes(t *testing.T) {
	m := loafergo.NewManager(nil)
	r := new(fake.Router)
	m.RegisterRoute(r)

	routes := m.GetRoutes()
	assert.Len(t, routes, 1)
	assert.Equal(t, r, routes[0])
}

func TestManager_RegisterRoutes(t *testing.T) {
	m := loafergo.NewManager(nil)
	r1 := new(fake.Router)
	r2 := new(fake.Router)

	m.RegisterRoutes([]loafergo.Router{r1, r2})

	routes := m.GetRoutes()
	assert.Len(t, routes, 2)
}

func TestManager_Run_StandardMessages(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	router := new(fake.Router)
	message := new(fake.Message)

	router.On("Configure", mock.Anything).Return(nil)
	router.On("WorkerPoolSize", mock.Anything).Return(int32(1))
	router.On("RunMode", mock.Anything).Return(loafergo.Parallel)
	router.On("GetMessages", mock.Anything).Return([]loafergo.Message{message}, nil).Once()
	router.On("GetMessages", mock.Anything).Return(nil, context.Canceled).Maybe()
	router.On("HandlerMessage", mock.Anything, message).Return(nil)
	router.On("Commit", mock.Anything, message).Return(nil)

	logger := new(fake.Logger)
	logger.On("Log", mock.Anything).Return()

	manager := loafergo.NewManager(&loafergo.Config{
		Logger:       logger,
		RetryTimeout: 1 * time.Second,
	})

	manager.RegisterRoute(router)

	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	err := manager.Run(ctx)
	assert.NoError(t, err)

	router.AssertExpectations(t)
}

func TestManager_Run_FIFO_With_GroupKey(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	message := new(fake.Message)
	message.On("SystemAttributeByKey", "MessageGroupId").Return("group1")
	message.On("Attribute", "seller_id").Return("123")

	router := new(fake.Router)
	router.On("Configure", mock.Anything).Return(nil)
	router.On("WorkerPoolSize", mock.Anything).Return(int32(1))
	router.On("RunMode", mock.Anything).Return(loafergo.PerGroupID)
	router.On("CustomGroupFields", mock.Anything).Return([]string{"seller_id"})
	router.On("GetMessages", mock.Anything).Return([]loafergo.Message{message}, nil).Once()
	router.On("GetMessages", mock.Anything).Return(nil, context.Canceled).Maybe()
	router.On("HandlerMessage", mock.Anything, message).Return(nil)
	router.On("Commit", mock.Anything, message).Return(nil)

	logger := new(fake.Logger)
	logger.On("Log", mock.Anything).Return()

	manager := loafergo.NewManager(&loafergo.Config{
		Logger:       logger,
		RetryTimeout: 1 * time.Second,
	})
	manager.RegisterRoute(router)

	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	err := manager.Run(ctx)
	assert.NoError(t, err)

	router.AssertExpectations(t)
	message.AssertExpectations(t)
}

func TestManager_Run_RouteConfigurationError(t *testing.T) {
	router := new(fake.Router)
	router.On("Configure", mock.Anything).Return(errors.New("config error"))

	logger := new(fake.Logger)
	logger.On("Log", mock.Anything, mock.Anything).Return()

	manager := loafergo.NewManager(&loafergo.Config{
		Logger: logger,
	})
	manager.RegisterRoute(router)

	err := manager.Run(context.Background())
	assert.EqualError(t, err, "config error")
}

func TestManager_Run_HandlerMessageError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	message := new(fake.Message)

	router := new(fake.Router)
	router.On("Configure", mock.Anything).Return(nil)
	router.On("WorkerPoolSize", mock.Anything).Return(int32(1))
	router.On("RunMode", mock.Anything).Return(loafergo.Parallel).Maybe()
	router.On("GetMessages", mock.Anything).Return([]loafergo.Message{message}, nil).Once()
	router.On("GetMessages", mock.Anything).Return(nil, context.Canceled).Maybe()
	router.On("HandlerMessage", mock.Anything, message).Return(errors.New("handler failed"))
	router.On("Commit", mock.Anything, message).Return(nil).Maybe()
	message.On("SystemAttributeByKey", "MessageGroupId").Return("group1").Maybe()
	message.On("Body").Return([]byte("body")).Maybe()
	message.On("Identifier").Return("id").Maybe()

	logger := new(fake.Logger)
	logger.On("Log", mock.Anything).Return()

	manager := loafergo.NewManager(&loafergo.Config{
		Logger:       logger,
		RetryTimeout: time.Second,
	})
	manager.RegisterRoute(router)

	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	err := manager.Run(ctx)
	assert.NoError(t, err)
	router.AssertExpectations(t)
}

func TestManager_Run_CommitError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	message := new(fake.Message)

	router := new(fake.Router)
	router.On("Configure", mock.Anything).Return(nil)
	router.On("WorkerPoolSize", mock.Anything).Return(int32(1))
	router.On("RunMode", mock.Anything).Return(loafergo.Parallel)
	router.On("GetMessages", mock.Anything).Return([]loafergo.Message{message}, nil).Once()
	router.On("GetMessages", mock.Anything).Return(nil, context.Canceled).Maybe()
	router.On("HandlerMessage", mock.Anything, message).Return(nil)
	router.On("Commit", mock.Anything, message).Return(errors.New("commit failed"))
	message.On("SystemAttributeByKey", "MessageGroupId").Return("group1").Maybe()
	message.On("Body").Return([]byte("body")).Maybe()
	message.On("Identifier").Return("id").Maybe()

	logger := new(fake.Logger)
	logger.On("Log", mock.Anything).Return()

	manager := loafergo.NewManager(&loafergo.Config{
		Logger:       logger,
		RetryTimeout: time.Second,
	})
	manager.RegisterRoute(router)

	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	err := manager.Run(ctx)
	assert.NoError(t, err)
	router.AssertExpectations(t)
}

func TestManager_Run_GetMessagesTemporaryError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	router := new(fake.Router)
	router.On("Configure", mock.Anything).Return(nil)
	router.On("WorkerPoolSize", mock.Anything).Return(int32(1))
	router.On("RunMode", mock.Anything).Return(loafergo.Parallel).Maybe()
	router.On("GetMessages", mock.Anything).Return(nil, errors.New("temporary error")).Once()
	router.On("GetMessages", mock.Anything).Return(nil, context.Canceled).Maybe()

	logger := new(fake.Logger)
	logger.On("Log", mock.Anything).Return()

	manager := loafergo.NewManager(&loafergo.Config{
		Logger:       logger,
		RetryTimeout: 50 * time.Millisecond,
	})
	manager.RegisterRoute(router)

	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	err := manager.Run(ctx)
	assert.NoError(t, err)
	router.AssertExpectations(t)
}
