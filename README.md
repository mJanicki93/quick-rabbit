## quick-rabbit

Library for creating services connected by RabbitMQ in fast way


Publisher service example:

```go
package main

import (
	"log"
	quick_rabbit "quick-rabbit"
)

type MyMessage struct {
	Name     string
	LastName string
	Age      int
}

func main() {
	rabbitChannel, err := quick_rabbit.NewRabbitFromURL("amqp://guest:guest@localhost:5672/", "services", nil)
	if err != nil {
		log.Println(err)
	}

	message := MyMessage{
		Name:     "John",
		LastName: "Doe",
		Age:      29,
	}

	rabbitChannel.PublishToQueue(message, "my-first-queue")
}

```

Consumer service example:
```go
package main

import (
	"github.com/streadway/amqp"
	quick_rabbit "quick-rabbit"
	"log"
)

type MyConsumer struct{}

func main() {
	rabbitChannel, err := quick_rabbit.NewRabbitFromURL("amqp://guest:guest@localhost:5672/", "services", nil)
	if err != nil {
		log.Println(err)
	}

	consumerEntity := MyConsumer{}
	err = rabbitChannel.RunConsumerOne(consumerEntity, "my-first-queue")
	if err != nil {
		log.Println(err)
	}

}

func (c MyConsumer) Consume(d amqp.Delivery) error {
	log.Printf("Received message: %v", d.Body)

	return nil
}
```
