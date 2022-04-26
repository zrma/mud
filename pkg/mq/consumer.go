package mq

import (
	"github.com/pkg/errors"
	"github.com/streadway/amqp"
)

type Consumer struct {
	ch    *amqp.Channel
	queue amqp.Queue
}

func (c Consumer) Consume() (<-chan amqp.Delivery, error) {
	message, err := c.ch.Consume(
		c.queue.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to register a consumer")
	}
	return message, nil
}
