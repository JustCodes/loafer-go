package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
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
	topicFifo    = "my_topic__test_f.fifo"
	topicOne     = "my_topic__test"
	topicTwo     = "my_topic__test2"
	queueOne     = "example-1"
	queueTwo     = "example-2"
	queueFifo    = "example.fifo"
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
	wg.Add(5)
	go produceStandardMessages(ctx, wg, producer, topicOne)
	go produceStandardMessages(ctx, wg, producer, topicTwo)
	go produceBatchMessages(ctx, wg, producer, topicOne)
	go produceBatchMessages(ctx, wg, producer, topicTwo)
	go produceFifoMessage(ctx, wg, producer, topicFifo)
	wg.Wait()
	log.Println("all messages were published")

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
				Handler:   handler1Standard,
				QueueName: queueOne,
			},
			sqs.RouteWithVisibilityTimeout(25),
			sqs.RouteWithMaxMessages(5),
			sqs.RouteWithWaitTimeSeconds(8),
			sqs.RouteWithWorkerPoolSize(workPool),
		),
		sqs.NewRoute(&sqs.Config{
			SQSClient: sqsClient,
			Handler:   handler2Standard,
			QueueName: queueTwo,
		}),
		sqs.NewRoute(
			&sqs.Config{
				SQSClient: sqsClient,
				Handler:   handler3Fifo,
				QueueName: queueFifo,
			},
			sqs.RouteWithRunMode(loafergo.PerGroupID),
			sqs.RouteWithCustomGroupFields([]string{"seller_id"}),
		),
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

func handler1Standard(ctx context.Context, m loafergo.Message) error {
	fmt.Printf("Message received handler1Standard:  %s\n ", string(m.Body()))
	return nil
}

func handler2Standard(ctx context.Context, m loafergo.Message) error {
	fmt.Printf("Message received handler2Standard: %s\n ", string(m.Body()))
	return nil
}

func handler3Fifo(ctx context.Context, m loafergo.Message) error {
	fmt.Printf("Message received handler3Fifo: %s\n ", string(m.Body()))
	return nil
}

func produceStandardMessages(ctx context.Context, wg *sync.WaitGroup, producer sns.Producer, topic string) {
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

func produceBatchMessages(ctx context.Context, wg *sync.WaitGroup, producer sns.Producer, topic string) {
	topicARN, err := sns.BuildTopicARN(awsRegion, awsAccountID, topic)
	if err != nil {
		log.Fatal(err)
	}

	input := &sns.PublishBatchInput{
		TopicARN: topicARN,
	}

	for i := 0; i < 10; i++ {
		input.Messages = append(input.Messages, &sns.PublishBatchEntry{
			ID:      fmt.Sprintf("id-%d", i),
			Message: fmt.Sprintf("{\"message\": \"Hello world! With batch mode!\", \"topic\": \"%s\", \"id\": %d}", topic, i),
		})
	}

	result, err := producer.ProduceBatch(ctx, input)
	if err != nil {
		log.Println("error to produce messages: ", err)
		return
	}

	outStr := ""
	for _, entry := range result.Successful {
		outStr += fmt.Sprintf("%s ", entry.MessageID)
	}

	fmt.Printf("Messages produced in batch to topic %s, IDs: %s\n", topic, outStr)

	wg.Done()
}

func produceFifoMessage(ctx context.Context, wg *sync.WaitGroup, producer sns.Producer, topic string) {
	for i := 0; i < 40; i++ {
		topicARN, err := sns.BuildTopicARN(awsRegion, awsAccountID, topic)
		if err != nil {
			log.Fatal(err)
		}

		sellerID := randInt(1, 3)
		id, err := producer.Produce(ctx, &sns.PublishInput{
			Attributes: map[string]string{
				"seller_id": fmt.Sprintf("%d", sellerID),
			},
			Message:         fmt.Sprintf("{\"message\": \"Hello world!\", \"topic\": \"%s\", \"id\": %d, \"seller_id\": %d}", topic, i+1, sellerID),
			GroupID:         "XPTO",
			DeduplicationID: fmt.Sprintf("%d", i+1),
			TopicARN:        topicARN,
		})
		if err != nil {
			log.Println("error to produce message: ", err)
			continue
		}
		fmt.Printf("Message produced to topic %s; id: %s \n", topic, id)
	}
	wg.Done()
}

func randInt(min int, max int) int {
	rand.NewSource(time.Now().UnixNano())
	return rand.Intn(max - min + 1)
}

func panicRecover() {
	if r := recover(); r != nil {
		log.Panicf("error: %v", r)
	}
	log.Println("example stopped")
}
