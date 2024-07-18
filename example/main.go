package main

import (
	"context"
	"fmt"

	loafergo "github.com/justcodes/loafer-go"
)

const (
	awsEndpoint = "http://localhost:4566"
	awsKey      = "dummy"
	awsSecret   = "dummy"
	awsRegion   = "us-east-1"
	awsProfile  = "test-profile"
	workPool    = 5
)

func main() {
	ctx := context.Background()
	c := &loafergo.Config{
		// for emulation only
		Hostname:   awsEndpoint,
		Key:        awsKey,
		Secret:     awsSecret,
		Region:     awsRegion,
		Profile:    awsProfile,
		WorkerPool: workPool,
	}
	manager := loafergo.NewManager(ctx, c)

	var routes = []*loafergo.Route{
		loafergo.NewRoute(
			"example-1",
			handler1,
			loafergo.RouteWithVisibilityTimeout(25),
			loafergo.RouteWithMaxMessages(5),
			loafergo.RouteWithWaitTimeSeconds(8),
		),
		loafergo.NewRoute("example-1", handler2),
	}

	manager.RegisterRoutes(routes)

	err := manager.Run()
	if err != nil {
		panic(err)
	}
}

func handler1(ctx context.Context, m loafergo.Message) error {
	fmt.Printf("Message received handler1: %+v\n ", m)
	return nil
}

func handler2(ctx context.Context, m loafergo.Message) error {
	fmt.Printf("Message received handler2: %+v\n ", m)
	return nil
}
