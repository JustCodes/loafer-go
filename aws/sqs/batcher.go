package sqs

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"

	loafergo "github.com/justcodes/loafer-go/v2"
)

const (
	// maxDeleteBatchSize is the AWS hard limit of entries accepted by DeleteMessageBatch.
	// It matches the ReceiveMessage MaxNumberOfMessages ceiling, so a full receive batch
	// collapses into a single delete call.
	maxDeleteBatchSize = 10
	// deleteRetryLimit is how many extra attempts a batch gets after a server side failure.
	deleteRetryLimit = 2
	// deleteRetryBackoff is the base delay between delete attempts.
	deleteRetryBackoff = 50 * time.Millisecond
	// flushTimeout bounds a flush, retries included. Flushes run detached from the route
	// context, so this is what keeps a hung request from stalling the batcher forever.
	flushTimeout = 5 * time.Second
	// messageGroupIDAttr is the FIFO group system attribute, used only for logging.
	messageGroupIDAttr = "MessageGroupId"
)

// deleteBatcher groups DeleteMessage calls into DeleteMessageBatch requests.
//
// Batching is opportunistic: there is no timer. The batcher blocks for the first pending
// delete, takes whatever else is already queued - up to 10 - and sends it right away. Under
// load the deletes committed while a request is in flight pile up and the next request goes
// out full; under light load a request carries a single entry, so no delete ever waits for
// another message and FIFO groups pay no added latency.
//
// A single goroutine owns the pending batch, so no locking is needed around it. The mutex
// only guards the closed flag, which makes enqueue safe against a concurrent shutdown.
type deleteBatcher struct {
	sqs      loafergo.SQSClient
	logger   loafergo.Logger
	in       chan loafergo.Message
	done     chan struct{}
	queueURL string
	mu       sync.RWMutex
	closed   bool
}

func newDeleteBatcher(
	client loafergo.SQSClient,
	logger loafergo.Logger,
	queueURL string,
	capacity int,
) *deleteBatcher {
	return &deleteBatcher{
		sqs:      client,
		logger:   logger,
		queueURL: queueURL,
		in:       make(chan loafergo.Message, capacity),
		done:     make(chan struct{}),
	}
}

// enqueue hands a message to the batcher. It never blocks: it reports false when the
// batcher is already shut down or its buffer is saturated, and the caller is then
// expected to fall back to a single DeleteMessage call.
func (b *deleteBatcher) enqueue(msg loafergo.Message) bool {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.closed {
		return false
	}

	select {
	case b.in <- msg:
		return true
	default:
		return false
	}
}

// run sends deletes until ctx is canceled. It must be started exactly once.
func (b *deleteBatcher) run(ctx context.Context) {
	defer close(b.done)

	batch := make([]loafergo.Message, 0, maxDeleteBatchSize)
	for {
		select {
		case <-ctx.Done():
			b.shutdown(ctx)
			return
		case msg := <-b.in:
			batch = b.collect(append(batch[:0], msg))
			b.flushDetached(ctx, batch)
		}
	}
}

// collect adds to batch whatever is already queued, without blocking, up to a full request.
func (b *deleteBatcher) collect(batch []loafergo.Message) []loafergo.Message {
	for len(batch) < maxDeleteBatchSize {
		select {
		case msg := <-b.in:
			batch = append(batch, msg)
		default:
			return batch
		}
	}
	return batch
}

// shutdown drains whatever is still queued and flushes it, so pending deletes are not
// silently lost.
func (b *deleteBatcher) shutdown(ctx context.Context) {
	b.mu.Lock()
	b.closed = true
	b.mu.Unlock()

	// Every enqueue either observed closed and fell back to a single delete, or already
	// landed in the buffer and is picked up here. There is no third case.
	if batch := b.drainAll(); len(batch) > 0 {
		b.flushDetached(ctx, batch)
	}
}

// drainAll takes every message already queued, without blocking.
func (b *deleteBatcher) drainAll() []loafergo.Message {
	var batch []loafergo.Message
	for {
		select {
		case msg := <-b.in:
			batch = append(batch, msg)
		default:
			return batch
		}
	}
}

// flushDetached flushes on a context detached from ctx: a delete already taken off the queue
// must not be aborted by the route shutting down, or the message is redelivered.
func (b *deleteBatcher) flushDetached(ctx context.Context, batch []loafergo.Message) {
	flushCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), flushTimeout)
	defer cancel()
	b.flushPending(flushCtx, batch)
}

// flushPending deletes every pending message, splitting it into requests of at most
// maxDeleteBatchSize entries.
func (b *deleteBatcher) flushPending(ctx context.Context, batch []loafergo.Message) {
	for len(batch) > 0 {
		n := len(batch)
		if n > maxDeleteBatchSize {
			n = maxDeleteBatchSize
		}
		b.flush(ctx, batch[:n])
		batch = batch[n:]
	}
}

// flush issues a single DeleteMessageBatch, retrying the entries that failed for server
// side reasons. Every message reaches a terminal state - and is dispatched - before it
// returns, so the visibility watchdog is never left running.
func (b *deleteBatcher) flush(ctx context.Context, batch []loafergo.Message) {
	pending := b.discardInvalid(batch)

	for attempt := 0; len(pending) > 0; attempt++ {
		if attempt > 0 {
			b.backoff(ctx, attempt)
		}
		pending = b.deleteOnce(ctx, pending, attempt >= deleteRetryLimit)
	}
}

// deleteOnce performs one delete round and returns the messages worth retrying.
// When last is set nothing is returned: every message is resolved and dispatched.
func (b *deleteBatcher) deleteOnce(ctx context.Context, pending []loafergo.Message, last bool) []loafergo.Message {
	out, err := b.sqs.DeleteMessageBatch(ctx, &sqs.DeleteMessageBatchInput{
		QueueUrl: &b.queueURL,
		Entries:  buildDeleteEntries(pending),
	})
	if err != nil {
		if !last {
			return pending
		}
		b.logger.Log(fmt.Sprintf(
			"delete_message_batch_error: %v; queueUrl: %s; count: %d",
			err, b.queueURL, len(pending),
		))
		dispatchAll(pending)
		return nil
	}

	return b.resolve(out, pending, last)
}

// resolve dispatches the messages the batch settled and returns the retryable ones.
func (b *deleteBatcher) resolve(
	out *sqs.DeleteMessageBatchOutput,
	pending []loafergo.Message,
	last bool,
) []loafergo.Message {
	failed := make(map[int]types.BatchResultErrorEntry, len(out.Failed))
	for _, f := range out.Failed {
		if f.Id == nil {
			continue
		}
		i, err := strconv.Atoi(*f.Id)
		if err != nil || i < 0 || i >= len(pending) {
			continue
		}
		failed[i] = f
	}

	var retry []loafergo.Message
	for i, msg := range pending {
		f, bad := failed[i]
		switch {
		case !bad:
			msg.Dispatch()
		case f.SenderFault || last:
			// A sender fault (an invalid receipt handle, for instance) will not heal on a
			// retry. The message is left to the queue visibility timeout, as it already is
			// today when DeleteMessage fails.
			b.logEntryError(msg, f)
			msg.Dispatch()
		default:
			retry = append(retry, msg)
		}
	}

	return retry
}

// discardInvalid drops messages without a receipt handle, which AWS would reject for the
// whole request, taking the valid entries down with them.
func (b *deleteBatcher) discardInvalid(batch []loafergo.Message) []loafergo.Message {
	valid := make([]loafergo.Message, 0, len(batch))
	for _, msg := range batch {
		if msg.Identifier() == "" {
			b.logger.Log(fmt.Sprintf(
				"delete_message_batch_skipped: empty receipt handle; queueUrl: %s; group_id: %s",
				b.queueURL, msg.SystemAttributeByKey(messageGroupIDAttr),
			))
			msg.Dispatch()
			continue
		}
		valid = append(valid, msg)
	}
	return valid
}

func (b *deleteBatcher) backoff(ctx context.Context, attempt int) {
	timer := time.NewTimer(time.Duration(attempt) * deleteRetryBackoff)
	defer timer.Stop()

	select {
	case <-timer.C:
	case <-ctx.Done():
	}
}

// logEntryError mirrors the commit_message_error log emitted by the manager when Commit
// is synchronous, so switching to batch deletes does not lose observability.
func (b *deleteBatcher) logEntryError(msg loafergo.Message, f types.BatchResultErrorEntry) {
	b.logger.Log(fmt.Sprintf(
		"commit_message_error: %s %s; sender_fault: %t; queueUrl: %s; message: %s; group_id: %s; identifier: %s",
		deref(f.Code), deref(f.Message), f.SenderFault, b.queueURL,
		msg.Body(), msg.SystemAttributeByKey(messageGroupIDAttr), msg.Identifier(),
	))
}

func buildDeleteEntries(pending []loafergo.Message) []types.DeleteMessageBatchRequestEntry {
	entries := make([]types.DeleteMessageBatchRequestEntry, 0, len(pending))
	for i, msg := range pending {
		// Ids only need to be unique within the request, and the index is what maps a
		// result entry back to its message - sqs.Message does not expose the MessageId.
		id := strconv.Itoa(i)
		handle := msg.Identifier()
		entries = append(entries, types.DeleteMessageBatchRequestEntry{
			Id:            &id,
			ReceiptHandle: &handle,
		})
	}
	return entries
}

func dispatchAll(msgs []loafergo.Message) {
	for _, msg := range msgs {
		msg.Dispatch()
	}
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
