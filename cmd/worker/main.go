package main

import (
	"context"
	"log"
	"os"

	"github.com/streadway/amqp"
	"google.golang.org/protobuf/encoding/protojson"

	"mud/pb"
	"mud/pkg/k8s"
	"mud/pkg/mq"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	secret, err := k8s.GetSecret()
	if err != nil {
		log.Fatal(err)
	}

	opt := mq.Option{
		Host:     "rabbitmq.rabbitmq.svc.cluster.local",
		Port:     5672,
		Id:       secret.Id,
		Password: secret.Password,
	}
	client, err := mq.New(opt)
	if err != nil {
		log.Fatalln(err)
	}
	defer client.Close()

	hostname, err := os.Hostname()
	if err != nil {
		log.Fatalln(err)
	}

	consumer, err := client.Consumer("task_queue")
	if err != nil {
		log.Fatalln(err)
	}

	if err := worker(ctx, consumer, hostname); err != nil {
		log.Fatalln(err)
	}
}

type Consumer interface {
	Consume() (<-chan amqp.Delivery, error)
}

func worker(ctx context.Context, consumer Consumer, hostname string) error {
	messages, err := consumer.Consume()
	if err != nil {
		return err
	}

	for ctx.Err() == nil {
		select {
		case <-ctx.Done():
			break
		case message := <-messages:
			body := message.Body
			msg := pb.MoveRequest{}
			err := protojson.Unmarshal(body, &msg)
			if err != nil {
				log.Println("failed", err)
			} else {
				log.Printf("[#%s] Received a message: %s to %s\n", hostname, msg.GetPlayer(), msg.GetDirection())
			}

			if err := message.Ack(false); err != nil {
				log.Println("Ack failed", err)
			} else {
				log.Println("Ack succeeded")
			}
		}
	}
	return nil
}
