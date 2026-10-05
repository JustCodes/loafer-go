//go:build integration

package sqs_test

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	loafergo "github.com/justcodes/loafer-go/v2"
	loafersqs "github.com/justcodes/loafer-go/v2/aws/sqs"
)

// deleteCountingClient wraps a real SQSClient and records how many delete calls of each
// shape reached the API, delegating everything else untouched.
type deleteCountingClient struct {
	loafergo.SQSClient
	mu            sync.Mutex
	singleDeletes int
	batchCalls    int
	batchEntries  int
}

func newDeleteCountingClient(c loafergo.SQSClient) *deleteCountingClient {
	return &deleteCountingClient{SQSClient: c}
}

func (c *deleteCountingClient) DeleteMessage(
	ctx context.Context,
	params *sqs.DeleteMessageInput,
	optFns ...func(*sqs.Options),
) (*sqs.DeleteMessageOutput, error) {
	c.mu.Lock()
	c.singleDeletes++
	c.mu.Unlock()
	return c.SQSClient.DeleteMessage(ctx, params, optFns...)
}

func (c *deleteCountingClient) DeleteMessageBatch(
	ctx context.Context,
	params *sqs.DeleteMessageBatchInput,
	optFns ...func(*sqs.Options),
) (*sqs.DeleteMessageBatchOutput, error) {
	c.mu.Lock()
	c.batchCalls++
	c.batchEntries += len(params.Entries)
	c.mu.Unlock()
	return c.SQSClient.DeleteMessageBatch(ctx, params, optFns...)
}

func (c *deleteCountingClient) stats() (single, calls, entries int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.singleDeletes, c.batchCalls, c.batchEntries
}

func newIntegrationClient(t *testing.T, ctx context.Context) *sqs.Client {
	t.Helper()

	awsCfg, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithRegion("us-east-1"),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("dummy", "dummy", "")),
	)
	require.NoError(t, err)
	awsCfg.BaseEndpoint = aws.String(envOr("AWS_ENDPOINT", "http://localhost:4566"))

	return sqs.NewFromConfig(awsCfg)
}

func createQueue(t *testing.T, ctx context.Context, c *sqs.Client, name string, attrs map[string]string) string {
	t.Helper()

	out, err := c.CreateQueue(ctx, &sqs.CreateQueueInput{QueueName: &name, Attributes: attrs})
	require.NoError(t, err)
	t.Cleanup(func() {
		_, _ = c.DeleteQueue(context.Background(), &sqs.DeleteQueueInput{QueueUrl: out.QueueUrl})
	})
	return *out.QueueUrl
}

func remainingMessages(t *testing.T, ctx context.Context, c *sqs.Client, queueURL string) string {
	t.Helper()

	out, err := c.GetQueueAttributes(ctx, &sqs.GetQueueAttributesInput{
		QueueUrl:       &queueURL,
		AttributeNames: []types.QueueAttributeName{"ApproximateNumberOfMessages", "ApproximateNumberOfMessagesNotVisible"},
	})
	require.NoError(t, err)
	return out.Attributes["ApproximateNumberOfMessages"] + "/" + out.Attributes["ApproximateNumberOfMessagesNotVisible"]
}

// TestIntegrationDeleteBatchAgainstLocalStack proves against the real SQS API surface that
// enabling RouteWithDeleteBatch collapses per-message deletes into batched ones, and that
// every message is actually removed from the queue.
//
// Requires LocalStack running locally:
//
//	docker compose -f docker-compose.integration.yml up -d
//	go test -tags=integration -race -run TestIntegrationDeleteBatch ./aws/sqs/... -v
func TestIntegrationDeleteBatchAgainstLocalStack(t *testing.T) {
	const messageCount = 100

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	rawClient := newIntegrationClient(t, ctx)
	queueName := fmt.Sprintf("delete-batch-test-%d", time.Now().UnixNano())
	queueURL := createQueue(t, ctx, rawClient, queueName, nil)

	for i := 0; i < messageCount; i++ {
		_, err := rawClient.SendMessage(ctx, &sqs.SendMessageInput{
			QueueUrl:    &queueURL,
			MessageBody: aws.String(fmt.Sprintf("message-%d", i)),
		})
		require.NoError(t, err)
	}

	countingClient := newDeleteCountingClient(rawClient)
	consumed := make(chan struct{}, messageCount)

	route := loafersqs.NewRoute(&loafersqs.Config{
		SQSClient: countingClient,
		QueueName: queueName,
		Handler: func(context.Context, loafergo.Message) error {
			consumed <- struct{}{}
			return nil
		},
	},
		loafersqs.RouteWithMaxMessages(10),
		loafersqs.RouteWithWaitTimeSeconds(1),
		loafersqs.RouteWithVisibilityTimeout(30),
		loafersqs.RouteWithDeleteBatch(),
	)

	manager := loafergo.NewManager(&loafergo.Config{Logger: loafergo.NoOpLogger{}})
	manager.RegisterRoute(route)

	managerDone := make(chan struct{})
	go func() {
		defer close(managerDone)
		_ = manager.Run(ctx)
	}()

	deadline := time.After(60 * time.Second)
	for i := 0; i < messageCount; i++ {
		select {
		case <-consumed:
		case <-deadline:
			t.Fatalf("only %d of %d messages were consumed", i, messageCount)
		}
	}

	// let the last batch settle before tearing the manager down
	require.Eventually(t, func() bool {
		_, _, entries := countingClient.stats()
		return entries >= messageCount
	}, 15*time.Second, 100*time.Millisecond, "not every message was deleted")

	cancel()
	<-managerDone

	single, calls, entries := countingClient.stats()
	t.Logf("single_deletes=%d batch_calls=%d batch_entries=%d remaining(visible/in-flight)=%s",
		single, calls, entries, remainingMessages(t, context.Background(), rawClient, queueURL))

	assert.Zero(t, single, "no message should be deleted one by one")
	assert.Equal(t, messageCount, entries, "every message must be deleted exactly once")
	// 100 messages in batches of 10 is the floor; allow slack for partial receive batches.
	assert.LessOrEqual(t, calls, messageCount/4,
		"batching must cut the number of delete calls well below one per message")
}

// TestIntegrationDeleteBatchFIFOAgainstLocalStack checks the case the batching is most
// likely to hurt: on a FIFO queue a pending delete holds back the whole message group.
// Ordering must survive, and the run must not slow to a crawl.
func TestIntegrationDeleteBatchFIFOAgainstLocalStack(t *testing.T) {
	const (
		groups         = 4
		perGroup       = 15
		messageCount   = groups * perGroup
		maxRunDuration = 60 * time.Second
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	rawClient := newIntegrationClient(t, ctx)
	queueName := fmt.Sprintf("delete-batch-fifo-test-%d.fifo", time.Now().UnixNano())
	queueURL := createQueue(t, ctx, rawClient, queueName, map[string]string{
		"FifoQueue":                 "true",
		"ContentBasedDeduplication": "true",
	})

	for g := 0; g < groups; g++ {
		for i := 0; i < perGroup; i++ {
			_, err := rawClient.SendMessage(ctx, &sqs.SendMessageInput{
				QueueUrl:       &queueURL,
				MessageBody:    aws.String(fmt.Sprintf("group-%d-message-%02d", g, i)),
				MessageGroupId: aws.String(fmt.Sprintf("group-%d", g)),
			})
			require.NoError(t, err)
		}
	}

	countingClient := newDeleteCountingClient(rawClient)

	var mu sync.Mutex
	received := map[string][]string{}
	consumed := make(chan struct{}, messageCount)

	route := loafersqs.NewRoute(&loafersqs.Config{
		SQSClient: countingClient,
		QueueName: queueName,
		Handler: func(_ context.Context, m loafergo.Message) error {
			group := m.SystemAttributeByKey("MessageGroupId")
			mu.Lock()
			received[group] = append(received[group], string(m.Body()))
			mu.Unlock()
			consumed <- struct{}{}
			return nil
		},
	},
		loafersqs.RouteWithMaxMessages(10),
		loafersqs.RouteWithWaitTimeSeconds(1),
		loafersqs.RouteWithVisibilityTimeout(30),
		loafersqs.RouteWithRunMode(loafergo.PerGroupID),
		loafersqs.RouteWithDeleteBatch(),
	)

	manager := loafergo.NewManager(&loafergo.Config{Logger: loafergo.NoOpLogger{}})
	manager.RegisterRoute(route)

	managerDone := make(chan struct{})
	start := time.Now()
	go func() {
		defer close(managerDone)
		_ = manager.Run(ctx)
	}()

	deadline := time.After(maxRunDuration)
	for i := 0; i < messageCount; i++ {
		select {
		case <-consumed:
		case <-deadline:
			t.Fatalf("only %d of %d messages were consumed in %s", i, messageCount, maxRunDuration)
		}
	}
	elapsed := time.Since(start)

	require.Eventually(t, func() bool {
		_, _, entries := countingClient.stats()
		return entries >= messageCount
	}, 15*time.Second, 100*time.Millisecond)

	cancel()
	<-managerDone

	single, calls, entries := countingClient.stats()
	t.Logf("fifo: elapsed=%s single_deletes=%d batch_calls=%d batch_entries=%d", elapsed, single, calls, entries)

	assert.Zero(t, single)
	assert.Equal(t, messageCount, entries)

	mu.Lock()
	defer mu.Unlock()
	require.Len(t, received, groups)
	for g := 0; g < groups; g++ {
		group := fmt.Sprintf("group-%d", g)
		want := make([]string, 0, perGroup)
		for i := 0; i < perGroup; i++ {
			want = append(want, fmt.Sprintf("group-%d-message-%02d", g, i))
		}
		assert.Equal(t, want, received[group], "FIFO ordering must survive batched deletes in %s", group)
	}
}
