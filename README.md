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
```

---

## 🔚 Acknowledgments

Inspired by:

- [loafer](https://github.com/georgeyk/loafer)
- [gosqs](https://github.com/qhenkart/gosqs)
