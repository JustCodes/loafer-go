package main

import (
	"context"
	"fmt"

	loafer_go "github.com/justcodes/loafer-go"
)

func main() {
	c := loafer_go.Config{
		// for emulation only
		Hostname:   "http://localhost:4100",
		Key:        "aws-key",
		Secret:     "aws-secret",
		Region:     "us-east-1",
		WorkerPool: 30,
	}
	manager := loafer_go.NewManager(c)

	var routes = []*loafer_go.Route{
		loafer_go.NewRoute("queuename-1", handler1, 10, 30, 10),
		loafer_go.NewRoute("queuename-2", handler2, 10, 30, 10),
	}

	manager.RegisterRoutes(routes)

	err := manager.Run()
	if err != nil {
		panic(err)
	}
}

func handler1(ctx context.Context, m loafer_go.Message) error {
	fmt.Printf("Message received handler1: %+v\n ", m)
	return nil
}

func handler2(ctx context.Context, m loafer_go.Message) error {
	fmt.Printf("Message received handler2: %+v\n ", m)
	return nil
}
