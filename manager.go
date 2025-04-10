package loafergo

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

const messageBufferFactor = 2

// Manager holds the routes and config fields
type Manager struct {
	config *Config
	routes []Router
}

// NewManager creates a new Manager with the given configuration
func NewManager(config *Config) *Manager {
	cfg := loadConfig(config)

	return &Manager{config: cfg}
}

// RegisterRoute register a new route to the Manager
func (m *Manager) RegisterRoute(route Router) {
	m.routes = append(m.routes, route)
}

// RegisterRoutes register more than one route to the Manager
func (m *Manager) RegisterRoutes(routes []Router) {
	m.routes = append(m.routes, routes...)
}

// Run the Manager distributing the worker pool by the number of routes
// returns errors if no routes
func (m *Manager) Run(ctx context.Context) error {
	if len(m.routes) == 0 {
		return ErrNoRoute
	}

	var wg sync.WaitGroup
	wg.Add(len(m.routes))

	for _, r := range m.routes {
		err := r.Configure(ctx)
		if err != nil {
			m.config.Logger.Log(err)
			return err
		}
		go func() {
			m.processRoute(ctx, r)
			wg.Done()
		}()
	}
	// wait for all routes to finish
	wg.Wait()
	return nil
}

func (m *Manager) processRoute(ctx context.Context, r Router) {
	size := r.WorkerPoolSize(ctx)
	message := make(chan Message, size*messageBufferFactor)
	defer close(message)

	for w := int32(1); w <= size; w++ {
		go m.worker(ctx, r, message)
	}

	m.config.Logger.Log("\nconsumers is ready to consume messages...")

	for {
		if errors.Is(ctx.Err(), context.Canceled) {
			m.config.Logger.Log("context canceled process route stopped")
			break
		}

		msgs, err := r.GetMessages(ctx)
		if err != nil {
			m.config.Logger.Log(
				fmt.Sprintf(
					"%s , retrying in %fs",
					ErrGetMessage.Context(err).Error(),
					m.config.RetryTimeout.Seconds(),
				),
			)
			time.Sleep(m.config.RetryTimeout)
			continue
		}

		for _, msg := range msgs {
			message <- msg
		}
	}
}

func (m *Manager) worker(ctx context.Context, r Router, msg <-chan Message) {
	for v := range msg {
		err := r.HandlerMessage(ctx, v)
		if err != nil {
			m.config.Logger.Log(err)
			continue
		}

		err = r.Commit(ctx, v)
		if err != nil {
			m.config.Logger.Log(err)
			continue
		}
	}
}

// GetRoutes returns the available routes as a slice of Router type
func (m *Manager) GetRoutes() []Router {
	return m.routes
}

func loadConfig(config *Config) *Config {
	cfg := &Config{
		Logger:       newDefaultLogger(),
		RetryTimeout: defaultRetryTimeout,
	}

	if config == nil {
		return cfg
	}

	if config.Logger != nil {
		cfg.Logger = config.Logger
	}

	if config.RetryTimeout > 0 {
		cfg.RetryTimeout = config.RetryTimeout
	}

	return cfg
}
