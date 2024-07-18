# Loafer Go Example

The code examples that show you how to create multiple handlers using `loafer go` to consume different queues within the same service

For it, we will use a Localstack running on a Docker container using Docker Compose.

> Make sure you have Docker and Docker Compose installed on your machine.

## Run local

### Running Localstack

From this directory execute the following Docker Compose command:

```sh
docker compose up -d
```

> All initial configs to `aws` are inside `./aws` folder.
> it will create two queues and produce some messages for each of them

More about Localstack you can find [here](https://github.com/localstack/localstack), including how to create the resources on localstack initialization.

### Run the example

When the container is ready, run the example with the command:

```shell
go run .
```

The output should be something like this:

```console
Message received handler1: &{Message:{Attributes:map[] Body:0xc000025d40 MD5OfBody:0xc000025d30 MD5OfMessageAttributes:<nil> MessageAttributes:map[] MessageId:0xc000025d20 ReceiptHandle:0xc000025d10 noSmithyDocumentSerde:{}} err:0xc000095500}
Message received handler1: &{Message:{Attributes:map[] Body:0xc000025d00 MD5OfBody:0xc000025cf0 MD5OfMessageAttributes:<nil> MessageAttributes:map[] MessageId:0xc000025ce0 ReceiptHandle:0xc000025cd0 noSmithyDocumentSerde:{}} err:0xc0000954a0}
Message received handler2: &{Message:{Attributes:map[] Body:0xc0002964d0 MD5OfBody:0xc0002964c0 MD5OfMessageAttributes:<nil> MessageAttributes:map[] MessageId:0xc0002964f0 ReceiptHandle:0xc0002964e0 noSmithyDocumentSerde:{}} err:0xc000095a40}
...
```
