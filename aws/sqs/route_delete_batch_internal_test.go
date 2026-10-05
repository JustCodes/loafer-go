package sqs

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsSqs "github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	loafergo "github.com/justcodes/loafer-go/v2"
	"github.com/justcodes/loafer-go/v2/fake"
)

// newTestRoute builds a configured route backed by a fake client, plus a canceled-on-cleanup
// context so the batcher goroutine never outlives the test.
func newTestRoute(t *testing.T, queueName string, optFns ...func(*RouteConfig)) (*route, *fake.SQSClient) {
	t.Helper()

	client := new(fake.SQSClient)
	client.On("GetQueueUrl", mock.Anything, mock.Anything).
		Return(&awsSqs.GetQueueUrlOutput{QueueUrl: aws.String(testQueueURL)}, nil)

	r, ok := NewRoute(&Config{
		SQSClient: client,
		QueueName: queueName,
		Handler:   func(context.Context, loafergo.Message) error { return nil },
	}, optFns...).(*route)
	require.True(t, ok)

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	require.NoError(t, r.Configure(ctx))
	t.Cleanup(func() {
		cancel()
		if r.batcher != nil {
			<-r.batcher.done
		}
	})

	return r, client
}

func receiveOutput(n int) *awsSqs.ReceiveMessageOutput {
	out := &awsSqs.ReceiveMessageOutput{}
	for i := 0; i < n; i++ {
		out.Messages = append(out.Messages, types.Message{
			ReceiptHandle: aws.String("handle-" + string(rune('a'+i))),
			Body:          aws.String(`{"Message":"body"}`),
			Attributes:    map[string]string{messageGroupIDAttr: "group-1"},
		})
	}
	return out
}

// Without the option the route must keep deleting one message per call, unchanged.
func TestRoute_DeleteBatchDisabledByDefault(t *testing.T) {
	r, client := newTestRoute(t, "test-queue")
	require.Nil(t, r.batcher)

	msg := testMessage("handle-a")
	client.On("DeleteMessage", mock.Anything, &awsSqs.DeleteMessageInput{
		QueueUrl:      aws.String(testQueueURL),
		ReceiptHandle: aws.String("handle-a"),
	}).Return(&awsSqs.DeleteMessageOutput{}, nil).Once()

	require.NoError(t, r.Commit(context.Background(), msg))
	client.AssertExpectations(t)
	client.AssertNotCalled(t, "DeleteMessageBatch", mock.Anything, mock.Anything)
	assert.True(t, dispatched(msg))
}

func TestRoute_DeleteBatchDisabledPropagatesTheDeleteError(t *testing.T) {
	r, client := newTestRoute(t, "test-queue")

	wantErr := errors.New("delete failed")
	client.On("DeleteMessage", mock.Anything, mock.Anything).
		Return(&awsSqs.DeleteMessageOutput{}, wantErr).Once()

	assert.ErrorIs(t, r.Commit(context.Background(), testMessage("handle-a")), wantErr)
}

// Commit must not block the worker and must not issue a single-message delete.
func TestRoute_CommitIsAsynchronousWhenBatchingIsEnabled(t *testing.T) {
	r, client := newTestRoute(t, "test-queue", RouteWithDeleteBatch())
	require.NotNil(t, r.batcher)

	client.On("DeleteMessageBatch", mock.Anything, mock.Anything).Return(
		func(_ context.Context, in *awsSqs.DeleteMessageBatchInput,
			_ ...func(*awsSqs.Options)) (*awsSqs.DeleteMessageBatchOutput, error) {
			return allSuccessful(in), nil
		})

	done := make(chan struct{})
	go func() {
		defer close(done)
		require.NoError(t, r.Commit(context.Background(), testMessage("handle-a")))
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Commit blocked while batching was enabled")
	}
	client.AssertNotCalled(t, "DeleteMessage", mock.Anything, mock.Anything)
}

// Each Commit is sent as soon as the batcher is free: a message never waits on another one
// of its receive, which is what keeps FIFO groups and slow siblings decoupled.
func TestRoute_CommitIsDeletedWithoutWaitingForTheRestOfTheReceive(t *testing.T) {
	r, client := newTestRoute(t, "test-queue", RouteWithDeleteBatch())

	client.On("ReceiveMessage", mock.Anything, mock.Anything).Return(receiveOutput(3), nil).Once()

	batches := make(chan []types.DeleteMessageBatchRequestEntry, 4)
	client.On("DeleteMessageBatch", mock.Anything, mock.Anything).Return(
		func(_ context.Context, in *awsSqs.DeleteMessageBatchInput,
			_ ...func(*awsSqs.Options)) (*awsSqs.DeleteMessageBatchOutput, error) {
			batches <- in.Entries
			return allSuccessful(in), nil
		})

	ctx := context.Background()
	msgs, err := r.GetMessages(ctx, loafergo.NoOpLogger{})
	require.NoError(t, err)
	require.Len(t, msgs, 3)

	// the other two are still being handled
	require.NoError(t, r.Commit(ctx, msgs[0]))

	select {
	case entries := <-batches:
		assert.Equal(t, []string{"handle-a"}, entryHandles(entries))
	case <-time.After(time.Second):
		t.Fatal("the delete waited for the rest of the receive")
	}
	client.AssertNotCalled(t, "DeleteMessage", mock.Anything, mock.Anything)

	// stop the visibility watchdog of the two messages left unhandled
	for _, m := range msgs[1:] {
		m.Dispatch()
	}
}

func TestRoute_BackedOffMessageIsNeverDeleted(t *testing.T) {
	r, client := newTestRoute(t, "test-queue", RouteWithDeleteBatch())

	msg := testMessage("handle-a")
	msg.backedOff = true

	require.NoError(t, r.Commit(context.Background(), msg))
	time.Sleep(80 * time.Millisecond)

	client.AssertNotCalled(t, "DeleteMessage", mock.Anything, mock.Anything)
	client.AssertNotCalled(t, "DeleteMessageBatch", mock.Anything, mock.Anything)
}

// Once the batcher is gone Commit must still delete the message, synchronously.
func TestRoute_FallsBackToSingleDeleteAfterShutdown(t *testing.T) {
	client := new(fake.SQSClient)
	client.On("GetQueueUrl", mock.Anything, mock.Anything).
		Return(&awsSqs.GetQueueUrlOutput{QueueUrl: aws.String(testQueueURL)}, nil)
	client.On("DeleteMessage", mock.Anything, mock.Anything).
		Return(&awsSqs.DeleteMessageOutput{}, nil).Once()

	r, ok := NewRoute(&Config{
		SQSClient: client,
		QueueName: "test-queue",
		Handler:   func(context.Context, loafergo.Message) error { return nil },
	}, RouteWithDeleteBatch()).(*route)
	require.True(t, ok)

	ctx, cancel := context.WithCancel(context.Background())
	require.NoError(t, r.Configure(ctx))
	cancel()
	<-r.batcher.done

	require.NoError(t, r.Commit(context.Background(), testMessage("handle-a")))
	client.AssertExpectations(t)
}

// Dispatch is called from the batcher, the fallback path and the handler error path; a
// second send on the buffered channel would deadlock a worker forever.
func TestMessage_DispatchIsIdempotent(t *testing.T) {
	m := testMessage("handle-a")

	done := make(chan struct{})
	go func() {
		defer close(done)
		m.Dispatch()
		m.Dispatch()
		m.Dispatch()
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("a repeated Dispatch blocked")
	}

	assert.True(t, dispatched(m))
	assert.False(t, dispatched(m), "the watchdog is signalled exactly once")
}

func TestMessage_IdentifierWithoutReceiptHandle(t *testing.T) {
	assert.Empty(t, testMessage("").Identifier())
}
