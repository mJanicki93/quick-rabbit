package quick_rabbit

import (
	"encoding/json"
	"log"
	"os"
	"os/signal"

	"github.com/streadway/amqp"
)

// A RabbitChannel stores all objects needed to
// work with RabbitMQ using amqp protocol.
type RabbitChannel struct {
	Connection *amqp.Connection
	Channel    *amqp.Channel
	Queue      amqp.Queue
	DeferFunc  func()
}

type RabbitConsumer interface {
	Consume(d amqp.Delivery) error
}

func (r RabbitChannel) RunConsumerOne(rabbitConsumerEntity RabbitConsumer) {
	err := r.ConsumeRabbitChannel(rabbitConsumerEntity.Consume)
	if err != nil {
		return
	}
}

func (r RabbitChannel) RunConsumerWithPrefetch(rabbitConsumerEntity RabbitConsumer, prefetchCount int) {
	err := r.Channel.Qos(prefetchCount, 0, false)
	if err != nil {
		log.Println(err)
		return
	}
	err = r.ConsumeRabbitChannel(rabbitConsumerEntity.Consume)
	if err != nil {
		log.Println(err)
		return
	}
}

func CreateRabbitChannelWithURL(queueName string, connectionURL string, exchange string) (RabbitChannel, error) {
	rabbitChannel := RabbitChannel{}
	conn, ch, deferFunc := getRabbitChannel(connectionURL, exchange)

	queue, err := ch.QueueDeclare(queueName, true, false, false, false, nil)
	if err != nil {
		return rabbitChannel, err
	}

	newDeferFunc := func() {
		deferFunc(conn)
	}

	rabbitChannel.DeferFunc = newDeferFunc
	rabbitChannel.Channel = ch
	rabbitChannel.Connection = conn
	rabbitChannel.Queue = queue

	return rabbitChannel, nil
}

func getRabbitChannel(connectionURL string, exchange string) (*amqp.Connection, *amqp.Channel, func(connection *amqp.Connection)) {
	conn, err := amqp.DialTLS(connectionURL, nil)
	if err != nil {
		log.Println(err)
		return nil, nil, nil
	}
	channel, err := conn.Channel()
	if err != nil {
		log.Println(err)
		return nil, nil, nil
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
		return nil, nil, nil
	}

	deferFunc := func(conn *amqp.Connection) {
		err := conn.Close()
		if err != nil {
			return
		}
	}
	return conn, channel, deferFunc
}

func (r RabbitChannel) PublishToRabbitChannel(v interface{}) {
	msg, _ := json.Marshal(v)

	err := r.Channel.QueueBind(r.Queue.Name, r.Queue.Name, "services", false, nil)
	if err != nil {
		log.Println(err)
	}

	pushToQueue(r.Channel, r.Queue.Name, msg)
}

func (r RabbitChannel) ConsumeRabbitChannel(consumeFunction func(delivery amqp.Delivery) error) error {
	msgs, err := r.Channel.Consume(
		r.Queue.Name, // queue
		"services",   // consumer
		false,        // auto-ack
		false,        // exclusive
		false,        // no-local
		false,        // no-wait
		nil,          // args
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
