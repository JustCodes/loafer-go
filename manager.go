package loafergo

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

const (
	defaultRetryTimeout = 10 * time.Second
)

// Manager holds the routes and config fields
type Manager struct {
	config *Config
	routes []Router
}

// NewManager creates a new Manager with the given configuration
func NewManager(config *Config) *Manager {
	if config == nil {
		config = loadDefaultConfig()
	}

	if config.Logger == nil {
		config.Logger = newDefaultLogger()
	}

	return &Manager{config: config}
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
	message := make(chan Message)
	defer close(message)

	for w := int32(1); w <= r.WorkerPoolSize(ctx); w++ {
		go m.worker(ctx, r, message)
	}

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
					defaultRetryTimeout.Seconds(),
				),
			)
			time.Sleep(defaultRetryTimeout)
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

func loadDefaultConfig() *Config {
	return &Config{}
}
