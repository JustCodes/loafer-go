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
> it will create topics, queues and subscriptions

More about Localstack you can find [here](https://github.com/localstack/localstack), including how to create the resources on localstack initialization.

### Run the example

When the container is ready, run the example with the command:

```shell
go run .
```

The output should be something like this:

```console
Message produced to topic my_topic__test2; id: 7750c733-abb8-4f37-81d4-202c505d5fec    
Message produced to topic my_topic__test; id: f8db976f-1631-43ce-9806-787196b5165b 
Message produced to topic my_topic__test; id: 373eb406-d4f3-415b-9729-fdba291c4f12 
Message produced to topic my_topic__test2; id: 60f5aed8-9429-42f8-ad2f-fd03e3d87a34 
Message produced to topic my_topic__test2; id: 10d7c4a9-88f5-4730-9328-8c2f110fd7a7 
...


******** Start queues consumers ********

Message received handler1:  {"Type": "Notification", "MessageId": "95aea4d0-55b2-41bf-b064-b474d2654272", "TopicArn": "arn:aws:sns:us-east-1:000000000000:my_topic__test", "Message": "{\"message\": \"Hello world!\", \"topic\": \"my_topic__test\", \"id\": 3}", "Timestamp": "2024-08-14T16:43:31.850Z", "UnsubscribeURL": "http://localhost.localstack.cloud:4566/?Action=Unsubscribe&SubscriptionArn=arn:aws:sns:us-east-1:000000000000:my_topic__test:8ccc2bbc-35a6-406b-b1e2-87edae9a46af", "SignatureVersion": "1", "Signature": "L7cixLw+OoZ64f3Ual7ryzNEQ2k7gIxC8KGJywXg3RaQvOp9Pb3/clvMoUU2C4QsPQnJmY9ymKu7aMx2ucE8Qlz6nzgad5Z9RsAfkU/IgFb24OPUcGa6s51az2sNBhDdKjo/O9yl/Rpal/YmSjbH2B8vEbhWD1Fg+GOhkpIpUvtDSOUsJtFXKjPeHIYTATJXD5+Ne93kUo7wH4WCNEiZdZfkvJhWJ7HNm3kHOe46NGTNJroMiXNsV3ZVrtiNxUZL3piinu94GRauV4PwT4HjmdE8N9MSVXSiZeEfakxTKf5G2O+uDCXN+IwslcU2Tfbccdt+3X410Ti1oHFulsr1Qg==", "SigningCertURL": "http://localhost.localstack.cloud:4566/_aws/sns/SimpleNotificationService-6c6f63616c737461636b69736e696365.pem"}
 Message received handler2: {"Type": "Notification", "MessageId": "10d7c4a9-88f5-4730-9328-8c2f110fd7a7", "TopicArn": "arn:aws:sns:us-east-1:000000000000:my_topic__test2", "Message": "{\"message\": \"Hello world!\", \"topic\": \"my_topic__test2\", \"id\": 2}", "Timestamp": "2024-08-14T16:43:31.822Z", "UnsubscribeURL": "http://localhost.localstack.cloud:4566/?Action=Unsubscribe&SubscriptionArn=arn:aws:sns:us-east-1:000000000000:my_topic__test2:ad9dd9ae-a18e-4398-9a10-94a67a96ce87", "SignatureVersion": "1", "Signature": "B7Hq+Byn15i9VnSphCqjvy0MspnibVjfwcvuf4f8hGa07G7QeeajUyWNEEAoFvfcgMRdkWJgtvZnV2uGd9rJrqhy0iERFY1UGUZl6un9Gkf8wfd+BFNdrJrz3VDDXuwQfy29k+kpvwLaYGMENzQwcddsqdjVFpy8/3PiavYjf5CApFavbcI4fWmZlLajJW1fIZDf5Qbsibs1QXvR6EI0re8v4wFkqNXylchA2YNyjYmgd6vsgvGqTF8wZ6uE7LLHJkpiTSwSSA5RYp2Ssbrx8PJPt8HTNu79jLVvuuSYjCrP9uETM4jD1XMYXTKHzK0kBvZdxP2yAdIhTIpyPSptlA==", "SigningCertURL": "http://localhost.localstack.cloud:4566/_aws/sns/SimpleNotificationService-6c6f63616c737461636b69736e696365.pem"}
...
```
