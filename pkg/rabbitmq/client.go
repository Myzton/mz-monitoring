package rabbitmq

import (
	"fmt"

	"github.com/rabbitmq/amqp091-go"
)

func InitRabbitMQ(url string) (*amqp091.Connection, *amqp091.Channel, error) {
	conn, err := amqp091.Dial(url)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, nil, fmt.Errorf("failed to open a channel: %w", err)
	}

	_, err = ch.QueueDeclare(
		"monitoring_tasks",
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, nil, fmt.Errorf("failed to declare a queue: %w", err)
	}

	return conn, ch, nil
}
