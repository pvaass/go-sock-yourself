package main

import (
	"encoding/json"
	"log"

	"github.com/streadway/amqp"
)

type ResponseWithUserID struct {
	UserID string
}

type AMQPListener struct {
	URL      string
	Exchange string
	output   *Output
	channel  *amqp.Channel
	conn     *amqp.Connection
}

func (listener *AMQPListener) SetURL(url string) {
	listener.URL = url
}

func (listener *AMQPListener) SetExchange(exchange string) {
	listener.Exchange = exchange
}

func (listener *AMQPListener) SetOutput(output *Output) {
	listener.output = output
}

func (listener *AMQPListener) Close() {
	listener.output.Info.Println("Closing amqp channel")
	listener.channel.Close()
	listener.conn.Close()
}

func (listener *AMQPListener) Init(registry *ChannelRegistry) {
	conn, err := amqp.Dial(listener.URL)
	if err != nil {
		listener.output.Info.Fatalln(err)
		return
	}
	listener.conn = conn

	channel, err := conn.Channel()
	if err != nil {
		listener.output.Info.Fatalln(err)
		return
	}

	err = channel.ExchangeDeclare(listener.Exchange, "fanout", true, false, false, false, amqp.Table{})
	if err != nil {
		listener.output.Info.Fatalln(err)
		return
	}
	listener.channel = channel

	q, err := channel.QueueDeclare(
		"testerino", // name
		false,       // durable
		false,       // delete when unused
		false,       // exclusive
		false,       // no-wait
		nil,         // arguments
	)
	if err != nil {
		log.Fatalf("%s", err)
	}

	msgs, err := channel.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		log.Fatalf("%s", err)
	}

	forever := make(chan bool)

	go func() {
		for d := range msgs {

			resp := Response{
				Body: string(d.Body),
			}

			var respWUserID ResponseWithUserID
			err := json.Unmarshal(d.Body, &respWUserID)
			if err != nil {
				log.Fatalf("%s", err)
			}

			channel := registry.Get(respWUserID.UserID)

			if channel != nil {
				channel <- resp
			}
		}
	}()

	<-forever
}
