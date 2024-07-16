package main

import (
	"context"
	"fmt"

	loafergo "github.com/justcodes/loafer-go"
)

func main() {
	c := &loafergo.Config{
		// for emulation only
		Hostname:   "http://localhost:4100",
		Key:        "aws-key",
		Secret:     "aws-secret",
		Region:     "us-east-1",
		WorkerPool: 30,
	}
	manager := loafergo.NewManager(c)

	var routes = []*loafergo.Route{
		loafergo.NewRoute("queuename-1", handler1, 10, 30, 10),
		loafergo.NewRoute("queuename-2", handler2, 10, 30, 10),
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
