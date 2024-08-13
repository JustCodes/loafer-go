package main

import (
	"context"
	"fmt"
	"log"

	loafergo "github.com/justcodes/loafer-go"
	"github.com/justcodes/loafer-go/aws"
	"github.com/justcodes/loafer-go/aws/sqs"
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
	defer panicRecover()
	ctx := context.Background()
	awsConfig := &aws.Config{
		Key:      awsKey,
		Secret:   awsSecret,
		Region:   awsRegion,
		Profile:  awsProfile,
		Hostname: awsEndpoint,
	}

	sqsClient, err := sqs.NewClient(ctx, &aws.ClientConfig{
		Config:     awsConfig,
		RetryCount: 4,
	})

	var routes = []loafergo.Router{
		sqs.NewRoute(
			&sqs.Config{
				SQSClient: sqsClient,
				Handler:   handler1,
				QueueName: "example-1",
			},
			sqs.RouteWithVisibilityTimeout(25),
			sqs.RouteWithMaxMessages(5),
			sqs.RouteWithWaitTimeSeconds(8),
			sqs.RouteWithWorkerPoolSize(workPool),
		),
		sqs.NewRoute(&sqs.Config{
			SQSClient: sqsClient,
			Handler:   handler2,
			QueueName: "example-2",
		}),
	}

	c := &loafergo.Config{}
	manager := loafergo.NewManager(c)
	manager.RegisterRoutes(routes)

	log.Println("starting consumers")
	err = manager.Run(ctx)
	if err != nil {
		panic(err)
	}
}

func handler1(ctx context.Context, m loafergo.Message) error {
	fmt.Printf("Message received handler1:  %s\n ", string(m.Body()))
	return nil
}

func handler2(ctx context.Context, m loafergo.Message) error {
	fmt.Printf("Message received handler2: %s\n ", string(m.Body()))
	return nil
}

func panicRecover() {
	if r := recover(); r != nil {
		log.Panicf("error: %v", r)
	}
	log.Println("example stopped")
}
