package mq

import (
	"mime"

	"github.com/pkg/errors"
	"github.com/streadway/amqp"
)

type Publisher struct {
	ch    *amqp.Channel
	queue amqp.Queue
}

func (p Publisher) Publish(body []byte) error {
	err := p.ch.Publish(
		"",
		p.queue.Name,
		false,
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  mime.TypeByExtension(".txt"),
			Body:         body,
		},
	)
	if err != nil {
		return errors.Wrap(err, "failed to publish a message")
	}
	return nil
}
