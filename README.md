# Loafer Go

---

Lib for Go with:

- Async pooling of AWS/SQS messages
- Producer of AWS/SNS messages

### Install

---

Manual install:

```bash
go get -u github.com/justcodes/loafer-go/v2
```

Golang import:

```go
import "github.com/justcodes/loafer-go/v2"
```

### Usage

---

```go
package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	loafergo "github.com/justcodes/loafer-go/v2"
	"github.com/justcodes/loafer-go/v2/aws"
	"github.com/justcodes/loafer-go/v2/aws/sns"
	"github.com/justcodes/loafer-go/v2/aws/sqs"
)

const (
	awsEndpoint  = "http://localhost:4566"
	awsKey       = "dummy"
	awsSecret    = "dummy"
	awsAccountID = "000000000000"
	awsRegion    = "us-east-1"
	awsProfile   = "test-profile"
	workPool     = 5
	topicOne     = "my_topic__test"
	topicTwo     = "my_topic__test2"
	queueOne     = "example-1"
	queueTwo     = "example-2"
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

	snsClient, err := sns.NewClient(ctx, &aws.ClientConfig{
		Config:     awsConfig,
		RetryCount: 4,
	})

	producer, err := sns.NewProducer(&sns.Config{
		SNSClient: snsClient,
	})
	if err != nil {
		log.Fatal(err)
	}

	// Produce message async
	wg := &sync.WaitGroup{}
	wg.Add(2)
	go produceMessage(ctx, wg, producer, topicOne)
	go produceMessage(ctx, wg, producer, topicTwo)
	wg.Wait()
	log.Println("all messages was published")

	log.Printf("\n\n******** Start queues consumers ********\n\n")
	time.Sleep(2 * time.Second)

	sqsClient, err := sqs.NewClient(ctx, &aws.ClientConfig{
		Config:     awsConfig,
		RetryCount: 4,
	})

	var routes = []loafergo.Router{
		sqs.NewRoute(
			&sqs.Config{
				SQSClient: sqsClient,
				Handler:   handler1,
				QueueName: queueOne,
			},
			sqs.RouteWithVisibilityTimeout(25),
			sqs.RouteWithMaxMessages(5),
			sqs.RouteWithWaitTimeSeconds(8),
			sqs.RouteWithWorkerPoolSize(workPool),
		),
		sqs.NewRoute(&sqs.Config{
			SQSClient: sqsClient,
			Handler:   handler2,
			QueueName: queueTwo,
		}),
		sqs.NewRoute(&sqs.Config{
			SQSClient: sqsClient,
			Handler:   handler3,
			QueueName: queueTwo,
		}),
	}

	c := &loafergo.Config{}
	manager := loafergo.NewManager(c)
	manager.RegisterRoutes(routes)

	// Run manager
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

func handler2(ctx context.Context, m loafergo.Message) error {
	fmt.Printf("Message received handler2: %s\n ", string(m.Body()))
	m.Backoff(2 * time.Hour) // visibility timeout will be extended in 2h
	return nil
}

func produceMessage(ctx context.Context, wg *sync.WaitGroup, producer sns.Producer, topic string) {
	for i := 0; i < 20; i++ {
		topicARN, err := sns.BuildTopicARN(awsRegion, awsAccountID, topic)
		if err != nil {
			log.Fatal(err)
		}

		id, err := producer.Produce(ctx, &sns.PublishInput{
			Message:  fmt.Sprintf("{\"message\": \"Hello world!\", \"topic\": \"%s\", \"id\": %d}", topic, i),
			TopicARN: topicARN,
		})
		if err != nil {
			log.Println("error to produce message: ", err)
			continue
		}
		fmt.Printf("Message produced to topic %s; id: %s \n", topic, id)
	}
	wg.Done()
}

func panicRecover() {
	if r := recover(); r != nil {
		log.Panicf("error: %v", r)
	}
	log.Println("example stopped")
}

```

### TODO

- [x] Add more tests
- [ ] Add support for sending messages to SQS
- [x] Add support for sending messages to SNS

### Acknowledgments

This lib is inspired by [loafer](https://github.com/georgeyk/loafer/) and [gosqs](https://github.com/qhenkart/gosqs).
