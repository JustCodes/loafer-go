package loafergo

import "sync"

// Manager holds the routes and config fields
type Manager struct {
	routes []*Route
	config Config
}

// NewManager creates a new Manager with the given configuration
func NewManager(config Config) *Manager { //nolint:gocritic
	return &Manager{config: config}
}

// RegisterRoute register a new route to the Manager
func (m *Manager) RegisterRoute(route *Route) {
	m.routes = append(m.routes, route)
}

// RegisterRoutes register more than one route to the Manager
func (m *Manager) RegisterRoutes(routes []*Route) {
	m.routes = append(m.routes, routes...)
}

// Run the Manager distributing the worker pool by the number of routes
func (m *Manager) Run() error {
	if len(m.routes) == 0 {
		return nil
	}
	// the worker pool is divided by the number of routes
	var workerPool = m.config.WorkerPool / len(m.routes)

	if workerPool == 0 {
		workerPool = 1
	}

	var wg sync.WaitGroup
	wg.Add(len(m.routes))

	for _, Route := range m.routes {
		err := Route.configure(&m.config)
		if err != nil {
			return err
		}
		r := Route
		go func() {
			r.run(workerPool)
			wg.Done()
		}()
	}
	// wait for all routes to finish
	wg.Wait()
	return nil
}
