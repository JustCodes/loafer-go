# 🧰 Loafer Go Example

This project demonstrates
how to consume multiple AWS SQS queues
using the [`loafer-go`](https://github.com/justcodes/loafer-go) library in a single service,
with support for **SNS topics** (standard and FIFO) and **SQS consumers** running concurrently.

It uses [LocalStack](https://github.com/localstack/localstack) to simulate AWS services locally via Docker Compose.

---

## ✅ Features

- 📦 Multiple SQS queue consumers in a single app
- 🔄 FIFO and Standard SNS topic publishing
- 🧵 Grouped parallel processing using `PerGroupID` mode
- 🌐 LocalStack-based environment for local AWS simulation
- ✨ Built-in retries, batching and message attributes

---

## 🚀 Getting Started

### 1. Requirements

- [Docker](https://docs.docker.com/get-docker/)
- [Docker Compose](https://docs.docker.com/compose/install/)
- [Go 1.19+](https://go.dev/doc/install)

---

### 2. Run LocalStack

Start the environment with preconfigured SNS topics and SQS queues:

```bash
docker compose up -d
```

> The script at `./aws/init-aws.sh` initializes topics, queues and subscriptions using AWS CLI + LocalStack.

---

### 3. Run the Example

When LocalStack is ready (check container logs), run the Go example:

```bash
go run .
```

---

## 📝 Example Output

You should see output similar to:

```console
Message produced to topic my_topic__test2; id: aed1...
Message produced to topic my_topic__test; id: dcc8...

Messages produced in batch to topic my_topic__test, IDs:
  4d4d36a3... fc3c5c84...

... 

******** Start queues consumers ********

Message received handler2Standard: {"Type": "Notification", "MessageId": "af6b2c43-7e3e-49c8-b282-33b34f2a6323", "TopicArn": "arn:aws:sns:us-east-1:000000000000:my_topic__test2", "Message": "{\"message\": \"Hello world!\", \"topic\": \"my_topic__test2\", \"id\": 19}", "Timestamp": "2025-05-15T19:31:08.915Z", "UnsubscribeURL": "http://localhost.localstack.cloud:4566/?Action=Unsubscribe&SubscriptionArn=arn:aws:sns:us-east-1:000000000000:my_topic__test2:d59c1a6d-8d5a-4afc-976d-1d7c3832ba97", "SignatureVersion": "1", "Signature": "pGw9mfmdku4JMmYowph07T8zlYRWVZkefwpps1NVy4ylRtywEaFak+4gE4ZYpgX368L+P1tJt2Hps3ICTxf8mH9eRK45HOA+9NCz+BHp8K1LTeBDa6dSx0ArLts5t0catgsfitCFFltYWO4go6je3QVVQACDqGfcB3H8TYMWBlmsvsZI0CebY5r+XrnN145RsfunI/R5lZIUNt/qtzzRa5r4mPq4uRxtGQVMV/KHD955puNkJlMNuTI4LTHlgNonB+nOR1zZP9jCeaAvorBSRZdpjApy7DaXQ9euULCDhuaqUMqwcwy61doCECbk2AGSE7c1wTicbP7LHjoUKfK03Q==", "SigningCertURL": "http://localhost.localstack.cloud:4566/_aws/sns/SimpleNotificationService-6c6f63616c737461636b69736e696365.pem"}
Message received handler3Fifo: {"Type": "Notification", "MessageId": "ff42f904-62b7-4261-85be-98f0a13331a8", "TopicArn": "arn:aws:sns:us-east-1:000000000000:my_topic__test_f.fifo", "Message": "{\"message\": \"Hello world!\", \"topic\": \"my_topic__test_f.fifo\", \"id\": 27, \"seller_id\": 0}", "Timestamp": "2025-05-15T19:31:09.066Z", "UnsubscribeURL": "http://localhost.localstack.cloud:4566/?Action=Unsubscribe&SubscriptionArn=arn:aws:sns:us-east-1:000000000000:my_topic__test_f.fifo:c72ec955-7d90-48f0-9ae7-3a1f8f86b7ef", "MessageAttributes": {"seller_id": {"Type": "String", "Value": "0"}}, "SequenceNumber": "15009515419264352282"}
Message received handler3Fifo: {"Type": "Notification", "MessageId": "9b0edb43-47fd-4efe-93be-8f54111c1e56", "TopicArn": "arn:aws:sns:us-east-1:000000000000:my_topic__test_f.fifo", "Message": "{\"message\": \"Hello world!\", \"topic\": \"my_topic__test_f.fifo\", \"id\": 28, \"seller_id\": 1}", "Timestamp": "2025-05-15T19:31:09.069Z", "UnsubscribeURL": "http://localhost.localstack.cloud:4566/?Action=Unsubscribe&SubscriptionArn=arn:aws:sns:us-east-1:000000000000:my_topic__test_f.fifo:c72ec955-7d90-48f0-9ae7-3a1f8f86b7ef", "MessageAttributes": {"seller_id": {"Type": "String", "Value": "1"}}, "SequenceNumber": "15009515419264352283"}
Message received handler3Fifo: {"Type": "Notification", "MessageId": "70c801eb-5bae-4427-87cb-ca1db79d6f2e", "TopicArn": "arn:aws:sns:us-east-1:000000000000:my_topic__test_f.fifo", "Message": "{\"message\": \"Hello world!\", \"topic\": \"my_topic__test_f.fifo\", \"id\": 29, \"seller_id\": 1}", "Timestamp": "2025-05-15T19:31:09.076Z", "UnsubscribeURL": "http://localhost.localstack.cloud:4566/?Action=Unsubscribe&SubscriptionArn=arn:aws:sns:us-east-1:000000000000:my_topic__test_f.fifo:c72ec955-7d90-48f0-9ae7-3a1f8f86b7ef", "MessageAttributes": {"seller_id": {"Type": "String", "Value": "1"}}, "SequenceNumber": "15009515419264352284"}
Message received handler3Fifo: {"Type": "Notification", "MessageId": "2a1922da-b03c-44d0-9cc4-1a21b0b27c8f", "TopicArn": "arn:aws:sns:us-east-1:000000000000:my_topic__test_f.fifo", "Message": "{\"message\": \"Hello world!\", \"topic\": \"my_topic__test_f.fifo\", \"id\": 30, \"seller_id\": 2}", "Timestamp": "2025-05-15T19:31:09.082Z", "UnsubscribeURL": "http://localhost.localstack.cloud:4566/?Action=Unsubscribe&SubscriptionArn=arn:aws:sns:us-east-1:000000000000:my_topic__test_f.fifo:c72ec955-7d90-48f0-9ae7-3a1f8f86b7ef", "MessageAttributes": {"seller_id": {"Type": "String", "Value": "2"}}, "SequenceNumber": "15009515419264352285"}

...
```

---

## 🧪 What's Happening

1. The app **publishes messages** to:
    - Standard SNS topics: `my_topic__test`, `my_topic__test2`
    - FIFO SNS topic: `my_topic__test_f.fifo`

2. Subscribed SQS queues receive the messages:
    - `example-1`, `example-2` for standard
    - `example.fifo` for FIFO (with deduplication + message group ID)

3. The consumer manager spins up workers per queue and **dispatches messages to handlers**:
    - `handler1Standard`, `handler2Standard`, `handler3Fifo`

4. FIFO processing uses:
   ```go
   loafergo.RouteWithRunMode(loafergo.PerGroupID),
   loafergo.RouteWithCustomGroupFields([]string{"seller_id"})
   ```

---

## 📂 Project Structure

```bash
.
├── aws
│   └── init-aws.sh       # SNS/SQS setup script
├── docker-compose.yml    # LocalStack environment
├── main.go               # Loafer Go example
└── README.md             # You're here :)
```

---

## 🛠️ Customization Tips

- To **add more queues**: create new `sqs.NewRoute` with handler
- To **add more SNS topics**: update `init-aws.sh`
- To **test error retries**: modify handler to simulate failure
- To **test FIFO grouping**: publish with `GroupID` and attributes

---

## 📚 References

- 🔗 [loafer-go documentation](https://github.com/justcodes/loafer-go)
- 🔗 [LocalStack docs](https://docs.localstack.cloud/)
- 🔗 [AWS SNS/SQS FIFO queues](https://docs.aws.amazon.com/AWSSimpleQueueService/latest/SQSDeveloperGuide/FIFO-queues.html)

---

## 🧼 Cleanup

To stop everything:

```bash
docker compose down -v
```

---

## 💬 Support

Open an issue or reach out at [github.com/justcodes/loafer-go](https://github.com/justcodes/loafer-go)

---
