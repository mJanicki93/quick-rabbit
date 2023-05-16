## quick-rabbit

Library for creating services connected by RabbitMQ in fast way


Publisher service example:

```go
package main

import (
	quick_rabbit "quick-rabbit"
)

type MyMessage struct {
	Name     string
	LastName string
	Age      int
}

func main() {
	rabbitChannel := quick_rabbit.NewRabbitWithURL("amqp://guest:guest@localhost:5672/", "services")

	message := MyMessage{
		Name:     "John",
		LastName: "Doe",
		Age:      29,
	}

	rabbitChannel.PushToQueue("my-firs-queue", message)
}

```

Consumer service example:
```go
package main

import (
	"github.com/streadway/amqp"
	quick_rabbit "quick-rabbit"
)

type MyConsumer struct{}

func main() {
	rabbitChannel := quick_rabbit.NewRabbitWithURL("amqp://guest:guest@localhost:5672/", "services")
	

	consumerEntity := MyConsumer{}
	rabbitChannel.RunConsumerOne(consumerEntity, "my-first-queue")

}

func (c MyConsumer) Consume(d amqp.Delivery) error {
	log.Printf("Received message: %v", d.Body)

	return nil
}
```
