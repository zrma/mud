package mq

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/streadway/amqp"
)

func New(option Option) (*Client, error) {
	endpoint := fmt.Sprintf(
		"amqp://%s:%s@%s:%d/",
		option.Id, option.Password, option.Host, option.Port,
	)
	conn, err := amqp.Dial(endpoint)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to connect to RabbitMQ")
	}

	ch, err := conn.Channel()
	if err != nil {
		func() { _ = conn.Close() }()
		return nil, errors.Wrap(err, "Failed to open a channel")
	}
	return &Client{conn: conn, ch: ch}, nil
}

type Option struct {
	Host     string
	Port     int
	Id       string
	Password string
}

type Client struct {
	conn *amqp.Connection
	ch   *amqp.Channel
}

func (c Client) Publisher(name string) (*Publisher, error) {
	queue, err := c.ch.QueueDeclare(
		name,
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to declare a queue")
	}
	return &Publisher{ch: c.ch, queue: queue}, nil
}

func (c Client) Consumer(name string) (*Consumer, error) {
	queue, err := c.ch.QueueDeclare(
		name,
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to declare a queue")
	}

	if err := c.ch.Qos(
		1,
		0,
		false,
	); err != nil {
		return nil, errors.Wrap(err, "failed to set QoS")
	}
	return &Consumer{ch: c.ch, queue: queue}, nil
}

func (c *Client) Close() {
	if c.ch != nil {
		_ = c.ch.Close()
		c.ch = nil
	}
	if c.conn != nil {
		_ = c.conn.Close()
		c.conn = nil
	}
}
