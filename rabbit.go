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
	Args       amqp.Table
	Queue      amqp.Queue
	DeferFunc  func()
}

func NewRabbitFromURL(url string, exchange string, args amqp.Table) (*Rabbit, error) {
	if args == nil {
		args := make(amqp.Table)
		args["x-message-ttl"] = int32(259200000)
	} else {
		if args["x-message-ttl"] == nil {
			args["x-message-ttl"] = int32(259200000)
		}
	}
	if err := args.Validate(); err != nil {
		return nil, err
	}
	conn, err := amqp.DialTLS(url, nil)
	if err != nil {
		return nil, err
	}
	channel, err := conn.Channel()
	if err != nil {
		return nil, err
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
		return nil, err
	}

	deferFunc := func() {
		err := conn.Close()
		if err != nil {
			return
		}
	}

	newRabbit := &Rabbit{Connection: conn, Channel: channel, Exchange: exchange, Args: args, DeferFunc: deferFunc}

	return newRabbit, nil
}

func (r *Rabbit) RunConsumerOne(rabbitConsumerEntity RabbitConsumer, queue string) error {
	err := r.consumeRabbitChannel(queue, rabbitConsumerEntity.Consume)
	if err != nil {
		return err
	}
	return nil
}

func (r *Rabbit) RunConsumerWithPrefetch(rabbitConsumerEntity RabbitConsumer, prefetchCount int, queue string) error {
	err := r.Channel.Qos(prefetchCount, 0, false)
	if err != nil {
		return err
	}
	err = r.consumeRabbitChannel(queue, rabbitConsumerEntity.Consume)
	if err != nil {
		return err
	}

	return nil
}

func (r *Rabbit) PublishToQueue(v interface{}, queue string) error {
	msg, _ := json.Marshal(v)

	_, err := r.Channel.QueueDeclare(queue, true, false, false, false, r.Args)
	if err != nil {
		return err
	}
	err = r.Channel.QueueBind(queue, queue, r.Exchange, false, nil)
	if err != nil {
		return err
	}

	err = pushToQueue(r.Channel, r.Exchange, queue, msg)
	if err != nil {
		return err
	}
	return nil
}

func (r *Rabbit) consumeRabbitChannel(queue string, consumeFunction func(delivery amqp.Delivery) error) error {

	_, err := r.Channel.QueueDeclare(queue, true, false, false, false, r.Args)
	if err != nil {
		log.Println(err)
	}
	err = r.Channel.QueueBind(queue, queue, r.Exchange, false, nil)
	if err != nil {
		log.Println(err)
	}

	msgs, err := r.Channel.Consume(
		queue,      // queue
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
			log.Printf("Received a message: %s", d.Body)
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

	log.Printf(" connected to %v", r.Queue)
	go func() {
		<-c
		os.Exit(1)
	}()
	<-forever

	return nil
}

func pushToQueue(ch *amqp.Channel, exchange string, queue string, data []byte) error {
	msg := amqp.Publishing{
		ContentType: "text/plain",
		Body:        data,
	}
	err := ch.Publish(exchange, queue, false, false, msg)
	if err != nil {
		return err
	}

	return nil
}
