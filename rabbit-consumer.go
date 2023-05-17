package quick_rabbit

import "github.com/streadway/amqp"

type RabbitConsumer interface {
	Consume(d amqp.Delivery) error
}
