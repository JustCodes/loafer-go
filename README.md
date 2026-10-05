# Loafer Go

Loafer Go is a lightweight Go library designed for high-throughput and concurrent processing of messages from AWS SQS queues and sending to AWS SNS topics.

---

## ✨ Features

- ✅ **Concurrent Message Consumers** with fixed worker pool size
- ✅ **FIFO Grouped Processing** 
  - Based `MessageGroupId` and custom fields (loafergo.PerGroupID)
  - Parallel (loafergo.Parallel)
- ✅ **SNS Producer** with support for both standard and FIFO topics
- ✅ **SQS Batch Receive and Parallel Handling**
- ✅ **Batched Deletes** (`sqs.RouteWithDeleteBatch`) — up to 10x fewer delete API calls
- ✅ **Simple API** with clean abstractions and interfaces
- ✅ **Test Coverage & Benchmarks**
- ✅ **Fully Configurable** via functional options

---

## 📦 Installation

```bash
go get -u github.com/justcodes/loafer-go/v2
```

Import into your project:

```go
import "github.com/justcodes/loafer-go/v2"
```

---

## 🚀 Quickstart Example

Start by writing a main application that produces messages to SNS and consumes from SQS.

[example](/example)

---

## 🐳 Local Development (with LocalStack)

Make sure you have Docker installed.

```bash
docker compose up -d
```

The init script in `./aws/init-aws.sh` will:

- Create topics (`standard` and `fifo`)
- Create queues
- Subscribe queues to the topics

---

## 🧪 Testing

Run tests:

```bash
make test
```

Run benchmarks:

```bash
make test-bench
```

Install formatters and linters:

```bash
make configure
```

---

## 📁 Project Structure

- `loafergo/` – Main package code
- `aws/` – AWS configuration, SQS/SNS clients, and route handlers
- `example/` – Sample producer and consumer demonstrating loafergo usage
- `fake/` – Fakes for tests

---

## 🧪 Benchmark (SNS & SQS)

```text
BenchmarkParserJSONToAnotherJSON_Small-12       12374347               474.9 ns/op           272 B/op          8 allocs/op
BenchmarkParserJSONToAnotherJSON_Medium-12       3689594              1573 ns/op             552 B/op         10 allocs/op
BenchmarkParserJSONToAnotherJSON_Large-12        1672772              3476 ns/op            1144 B/op          7 allocs/op
BenchmarkStructToMap-12                          4388439              1468 ns/op            1160 B/op          8 allocs/op
BenchmarkStructToMapConcurrent-12                4082488              1512 ns/op            1160 B/op          8 allocs/op
BenchmarkStructToMapWithCache-12                 6046621              1025 ns/op            1160 B/op          8 allocs/op
```

---

## 🗑️ Batched Deletes

By default each consumed message costs one `DeleteMessage` call. `sqs.RouteWithDeleteBatch`
groups them into `DeleteMessageBatch` requests of up to 10 entries — the same ceiling as
`MaxNumberOfMessages`:

```go
route := sqs.NewRoute(
    &sqs.Config{SQSClient: client, Handler: handler, QueueName: "example-1"},
    sqs.RouteWithMaxMessages(10),
    sqs.RouteWithDeleteBatch(),
    sqs.RouteWithLogger(logger), // where delete failures are reported
)
```

Batching is opportunistic, with no timer: a delete goes out as soon as the batcher is free,
together with whatever else was committed while the previous request was in flight. Under
load requests fill up to 10 entries; under light load they carry a single one, so a delete
never waits on another message and costs no more than `DeleteMessage` did. Measured against
LocalStack with 2000 messages: 2000 `DeleteMessage` calls become 205 `DeleteMessageBatch`
calls (81.5% fewer billed calls overall), and the run is 2x faster since workers no longer
block on a delete round trip.

Two things to know before enabling it:

- **`Commit` becomes asynchronous.** It hands the message to the batcher and returns `nil`
  immediately, so the delete outcome is reported through the route logger instead of being
  returned to the manager.
- **Handlers must be idempotent.** A delete still pending when the process dies means the
  message is redelivered — which is already true of SQS at-least-once delivery, but the
  window is slightly wider.

---

## 📌 Makefile Tasks

```makefile
make update-dependencies     # Update Go dependencies
make format                  # Run goimports
make lint                    # Run golangci-lint
make install-golang-ci       # Install GolangCI-Lint
make install-goimports       # Install GoImports
make clean                   # Clean test cache
make test                    # Run tests with coverage
make test-bench              # Run benchmarks
make test-integration        # Run integration tests against LocalStack
make generate                # Regenerate the mocks in fake/
```

---

## 🔚 Acknowledgments

Inspired by:

- [loafer](https://github.com/georgeyk/loafer)
- [gosqs](https://github.com/qhenkart/gosqs)
