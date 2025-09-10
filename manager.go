package loafergo

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

const messageGroupID = "MessageGroupId"

// Manager coordinates multiple routes and startWorker pools.
type Manager struct {
	config *Config
	routes []Router
}

// NewManager creates a new Manager with the provided configuration.
func NewManager(config *Config) *Manager {
	return &Manager{
		config: loadConfig(config),
	}
}

// RegisterRoute adds a single route to the manager.
func (m *Manager) RegisterRoute(route Router) {
	m.routes = append(m.routes, route)
}

// RegisterRoutes adds multiple routes to the manager.
func (m *Manager) RegisterRoutes(routes []Router) {
	m.routes = append(m.routes, routes...)
}

// GetRoutes returns all registered routes.
func (m *Manager) GetRoutes() []Router {
	return m.routes
}

// Run the Manager distributing the startWorker pool by the number of routes.
// Returns an error if no routes are registered.
func (m *Manager) Run(ctx context.Context) error {
	if len(m.routes) == 0 {
		return ErrNoRoute
	}

	var wg sync.WaitGroup
	wg.Add(len(m.routes))

	for _, r := range m.routes {
		route := r // avoid closure over loop variable

		if err := route.Configure(ctx); err != nil {
			m.config.Logger.Log("route configuration failed:", err)
			return err
		}

		go func() {
			defer wg.Done()
			m.runRoute(ctx, route)
		}()
	}

	wg.Wait()
	return nil
}

func (m *Manager) runRoute(ctx context.Context, r Router) {
	workerCount := int(r.WorkerPoolSize(ctx))
	messageChs := make([]chan Message, workerCount)

	for i := 0; i < workerCount; i++ {
		messageChs[i] = make(chan Message)
		go m.startWorker(ctx, r, messageChs[i])
	}

	defer func() {
		for _, ch := range messageChs {
			close(ch)
		}
	}()

	m.config.Logger.Log("Route consumer ready...")

	for {
		select {
		case <-ctx.Done():
			m.config.Logger.Log("Context canceled; shutting down route.")
			return
		default:
			msgs, err := r.GetMessages(ctx, m.config.Logger)
			if err != nil {
				m.config.Logger.Log(fmt.Sprintf("%s, retrying in %.2fs", ErrGetMessage.Context(err).Error(), m.config.RetryTimeout.Seconds()))
				select {
				case <-ctx.Done():
					return
				case <-time.After(m.config.RetryTimeout):
					continue
				}
			}

			for _, msg := range msgs {
				index := m.assignWorkerIndex(ctx, msg, r, workerCount)
				select {
				case messageChs[index] <- msg:
				case <-ctx.Done():
					m.config.Logger.Log("Context done; shutting down route.")
					return
				}
			}
		}
	}
}

func (m *Manager) assignWorkerIndex(ctx context.Context, msg Message, r Router, size int) int {
	if r.RunMode(ctx) == PerGroupID {
		key := m.buildGroupKey(ctx, msg, r)
		return hashGroupID(key) % size
	}
	return rand.Intn(size)
}

func (m *Manager) startWorker(ctx context.Context, r Router, msgCh <-chan Message) {
	for msg := range msgCh {
		if err := r.HandlerMessage(ctx, msg); err != nil {
			logMsg := fmt.Sprintf(
				"handler_message_error: %v; message: %s; group_id: %s; identifier: %s",
				err, msg.Body(), msg.SystemAttributeByKey(messageGroupID), msg.Identifier(),
			)
			m.config.Logger.Log(logMsg)
			continue
		}
		if err := r.Commit(ctx, msg); err != nil {
			logMsg := fmt.Sprintf(
				"commit_message_error: %v; message: %s; group_id: %s; identifier: %s",
				err, msg.Body(), msg.SystemAttributeByKey(messageGroupID), msg.Identifier(),
			)
			m.config.Logger.Log(logMsg)
		}
	}
}

func (m *Manager) buildGroupKey(ctx context.Context, msg Message, r Router) string {
	key := msg.SystemAttributeByKey(messageGroupID)
	for _, field := range r.CustomGroupFields(ctx) {
		if v := msg.Attribute(field); v != "" {
			key += ":" + v
		}
	}
	return key
}

func hashGroupID(s string) int {
	h := 0
	for _, c := range s {
		h = int(c) + ((h << 5) - h)
	}
	return abs(h)
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}
