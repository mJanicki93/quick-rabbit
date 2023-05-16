## quick-rabbit

Library for creating services connected by RabbitMQ in fast way


Publisher service example:
```go
package main

import (
	"log"
)

type MyMessage struct {
	Name     string
	LastName string
	Age      int
}

func main() {
	rabbitChannel, err := CreateRabbitChannelWithURL("myFirstQueue", "amqp://guest:guest@localhost:5672/", "services")
	if err != nil {
		log.Println(err)
	}

	message := MyMessage{
		Name:     "John",
		LastName: "Doe",
		Age:      29,
	}

	rabbitChannel.PublishToRabbitChannel(message)
}

```

Consumer service example:
```go
package main

import (
	"github.com/streadway/amqp"
	"log"
)

type MyConsumer struct{}

func main() {
	rabbitChannel, err := CreateRabbitChannelWithURL("myFirstQueue", "amqp://guest:guest@localhost:5672/", "services")
	if err != nil {
		log.Println(err)
	}

	consumerEntity := MyConsumer{}
	rabbitChannel.RunConsumerOne(consumerEntity)

}

func (c MyConsumer) Consume(d amqp.Delivery) error {
	log.Printf("Received message: %v", d.Body)

	return nil
}
```
