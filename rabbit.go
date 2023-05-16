package quick_rabbit

import (
	"encoding/json"
	"log"
	"os"
	"os/signal"

	"github.com/streadway/amqp"
)

// A Rabbit stores all objects needed to
// work with RabbitMQ using amqp protocol.
type Rabbit struct {
	Connection *amqp.Connection
	Channel    *amqp.Channel
	Exchange   string
	Queue      amqp.Queue
	DeferFunc  func()
}

func NewRabbitWithURL(url string, exchange string) *Rabbit {
	conn, err := amqp.DialTLS(url, nil)
	if err != nil {
		log.Println(err)
		return nil
	}
	channel, err := conn.Channel()
	if err != nil {
		log.Println(err)
		return nil
	}
	deferFunc := func(conn *amqp.Connection) {
		err := conn.Close()
		if err != nil {
			log.Println(err)
			return
		}
	}

	err = channel.ExchangeDeclare(
		exchange,
		"direct",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil
	}

	newDeferFunc := func() {
		deferFunc(conn)
	}
	newRabbit := &Rabbit{Connection: conn, Channel: channel, DeferFunc: newDeferFunc, Exchange: exchange}

	return newRabbit
}

func (r *Rabbit) PushToQueue(queueName string, v interface{}) {
	msg, _ := json.Marshal(v)

	_, err := r.Channel.QueueDeclare(queueName, true, false, false, false, nil)
	if err != nil {
		log.Println(err)
	}

	err = r.Channel.QueueBind(queueName, queueName, r.Exchange, false, nil)
	if err != nil {
		log.Println(err)
	}

	_, err = r.Channel.QueueDeclare(queueName, true, false, false, false, nil)
	if err != nil {
		log.Println(err)
	}

	pushToQueue(r.Channel, r.Queue.Name, msg)
}

func (r *Rabbit) RunConsumerOne(rabbitConsumerEntity RabbitConsumer, queueName string) {
	err := r.consumeRabbitQueue(queueName, rabbitConsumerEntity.Consume)
	if err != nil {
		return
	}
}

func (r *Rabbit) RunConsumerWithPrefetch(rabbitConsumerEntity RabbitConsumer, prefetchCount int, queueName string) {
	err := r.Channel.Qos(prefetchCount, 0, false)
	if err != nil {
		log.Println(err)
		return
	}
	err = r.consumeRabbitQueue(queueName, rabbitConsumerEntity.Consume)
	if err != nil {
		log.Println(err)
		return
	}
}

func (r *Rabbit) consumeRabbitQueue(queueName string, consumeFunction func(delivery amqp.Delivery) error) error {
	msgs, err := r.Channel.Consume(
		queueName,  // queue
		r.Exchange, // consumer
		false,      // auto-ack
		false,      // exclusive
		false,      // no-local
		false,      // no-wait
		nil,        // args
	)
	if err != nil {
		return err
	}

	var forever chan struct{}
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	go func() {
		for d := range msgs {
			err := consumeFunction(d)
			if err != nil {
				log.Println(err)
				continue
			}
			err = d.Ack(false)
			if err != nil {
				log.Println(err)
			}
		}
	}()

	log.Printf(" connected to %v", queueName)
	go func() {
		<-c
		os.Exit(1)
	}()
	<-forever

	return nil
}

func pushToQueue(ch *amqp.Channel, queue string, data []byte) {
	msg := amqp.Publishing{
		ContentType: "text/plain",
		Body:        data,
	}
	err := ch.Publish("services", queue, false, false, msg)
	if err != nil {
		log.Println(err)
		panic(err)
	}
}
