package loafergo

import (
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
)

// Manager holds the routes and config fields
type Manager struct {
	routes []*Route
	config *Config
}

// NewManager creates a new Manager with the given configuration
func NewManager(config *Config) *Manager {
	if config.Logger == nil {
		config.Logger = newDefaultLogger()
	}
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

	for _, r := range m.routes {
		s, err := m.newSession()
		if err != nil {
			return err
		}
		err = r.configure(s, m.config.Logger)
		if err != nil {
			return err
		}
		go func() {
			r.run(workerPool)
			wg.Done()
		}()
	}
	// wait for all routes to finish
	wg.Wait()
	return nil
}

func (m *Manager) newSession() (*session.Session, error) {
	c := credentials.NewStaticCredentials(m.config.Key, m.config.Secret, "")
	_, err := c.Get()
	if err != nil {
		return nil, ErrInvalidCreds.Context(err)
	}

	r := &retryer{retryCount: m.config.RetryCount}

	cfg := request.WithRetryer(aws.NewConfig().WithRegion(m.config.Region).WithCredentials(c), r)

	// if an optional hostname config is provided, then replace the default one
	//
	// This will set the default AWS URL to a hostname of your choice. Perfect for testing, or mocking functionality
	if m.config.Hostname != "" {
		cfg.Endpoint = &m.config.Hostname
	}

	return session.NewSession(cfg)
}
