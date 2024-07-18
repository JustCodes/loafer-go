package loafergo

import (
	"context"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
)

const defaultMaxRetries = 10

// Manager holds the routes and config fields
type Manager struct {
	routes []*Route
	config *Config
	ctx    context.Context
}

// NewManager creates a new Manager with the given configuration
func NewManager(ctx context.Context, config *Config) *Manager {
	if config.Logger == nil {
		config.Logger = newDefaultLogger()
	}

	if config.RetryCount == 0 {
		config.RetryCount = defaultMaxRetries
	}

	return &Manager{config: config, ctx: ctx}
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
		err = r.configure(m.ctx, s, m.config.Logger)
		if err != nil {
			return err
		}
		go func() {
			r.run(m.ctx, workerPool)
			wg.Done()
		}()
	}
	// wait for all routes to finish
	wg.Wait()
	return nil
}

func (m *Manager) newSession() (cfg aws.Config, err error) {
	c := aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider(m.config.Key, m.config.Secret, ""))
	_, err = c.Retrieve(m.ctx)
	if err != nil {
		return cfg, ErrInvalidCreds.Context(err)
	}

	conf := []func(*config.LoadOptions) error{
		config.WithRegion(m.config.Region),
		config.WithCredentialsProvider(c),
		config.WithRetryMaxAttempts(m.config.RetryCount),
	}

	if m.config.Profile != "" {
		conf = append(conf, config.WithSharedConfigProfile(m.config.Profile))
	}

	cfg, err = config.LoadDefaultConfig(
		m.ctx,
		conf...,
	)

	// if an optional hostname config is provided, then replace the default one
	//
	// This will set the default AWS URL to a hostname of your choice. Perfect for testing, or mocking functionality
	if m.config.Hostname != "" {
		cfg.BaseEndpoint = aws.String(m.config.Hostname)
	}

	if err != nil {
		return
	}

	return
}
