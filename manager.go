package loafer_go

import "sync"

type manager struct {
	routes []*Route
	config Config
}

// NewManager creates a new manager with the given configuration
func NewManager(config Config) *manager {
	return &manager{config: config}
}

// Register a Route to the manager
func (m *manager) RegisterRoute(Route *Route) {
	m.routes = append(m.routes, Route)
}

// Register Routes to the manager
func (m *manager) RegisterRoutes(routes []*Route) {
	m.routes = append(m.routes, routes...)
}

// Run the manager distributing the worker pool by the number of routes
func (m *manager) Run() error {
	if len(m.routes) == 0 {
		return nil
	}
	// the worker pool is divided by the number of routes
	var workerPool int = m.config.WorkerPool / len(m.routes)

	if workerPool == 0 {
		workerPool = 1
	}

	var wg sync.WaitGroup
	wg.Add(len(m.routes))

	for _, Route := range m.routes {
		err := Route.configure(m.config)
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
