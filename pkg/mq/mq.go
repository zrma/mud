package mq

import (
	"errors"
	"fmt"

	"github.com/streadway/amqp"
)

func New(option Option) (*Wrapper, error) {
	endpoint := fmt.Sprintf(
		"amqp://%s:%s@%s:%d/",
		option.Id, option.Password, option.Host, option.Port,
	)
	conn, err := amqp.Dial(endpoint)
	if err != nil {
		return nil, errMsg(err, "Failed to connect to RabbitMQ")
	}

	ch, err := conn.Channel()
	if err != nil {
		func() { _ = conn.Close() }()
		return nil, errMsg(err, "Failed to open a channel")
	}

	return &Wrapper{Conn: conn, Chan: ch}, nil
}

func errMsg(err error, msg string) error {
	return errors.New(fmt.Sprintf("%s: %s", msg, err))
}

type Option struct {
	Host     string
	Port     int
	Id       string
	Password string
}

type Wrapper struct {
	Conn *amqp.Connection
	Chan *amqp.Channel
}

func (w *Wrapper) Close() {
	if w.Chan != nil {
		_ = w.Chan.Close()
		w.Chan = nil
	}
	if w.Conn != nil {
		_ = w.Conn.Close()
		w.Conn = nil
	}
}
