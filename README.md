# loafer-go
## Lib for GO with async pooling of AWS/SQS messages

### Usage
```go   
import (
  "context"
  "fmt"
  
  loafer_go "github.com/justcodes/loafer-go"
)

func main() {
  c := loafer_go.Config{
  // The hostname is used only when using localstack/goaws
  Hostname:   "http://localhost:4100",
  Key:        "aws-key",
  Secret:     "aws-secret",
  Region:     "us-east-1",
  // the number of workers will be divided between the routes
  // in this example, each route will have 15 workers
  // you should use with care, because using too many workers 
  // can cause problems consuming too much resources
  WorkerPool: 30, 
  }

  // initialize the manager
  manager := loafer_go.NewManager(c)

  // there are two ways to register routes
  // this way you can register various routes in a single call to 
  // manager.RegisterRoutes
  var routes = []*loafer_go.Route{
    loafer_go.NewRoute("queuename-1", handler1, 10, 30, 10),
    loafer_go.NewRoute("queuename-2", handler2, 10, 30, 10),
  }

  manager.RegisterRoutes(routes)

  // or you can register various routes calling the manager.RegisterRoute
  // method multiple times
  manager.RegisterRoute(loafer_go.NewRoute("queuename-1", handler1, 10, 30, 10))
  manager.RegisterRoute(loafer_go.NewRoute("queuename-2", handler1, 10, 30, 10))

  // start the manager it will run until you stop it with Ctrl + C
  err := manager.Run()

  if err != nil {
    panic(err)
  }
}

func handler1(ctx context.Context, m loafer_go.Message) error {
  fmt.Printf("Message received handler1: %+v\n ", m)
  // you can return errors, if you return an error the message will be returned to the queue
  return nil
}

func handler2(ctx context.Context, m loafer_go.Message) error {
  fmt.Printf("Message received handler2: %+v\n ", m)
  return nil
}
```

### TODO
- [ ] Add tests
- [ ] Add support for sending messages to SQS
- [ ] Add support for sending messages to SNS



### Acknowledgments

This lib is inspired by [loafer](https://github.com/georgeyk/loafer/) and [gosqs](https://github.com/qhenkart/gosqs).

I used [goaws](https://github.com/p4tin/goaws) for testing.