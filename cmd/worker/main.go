package main

import (
	"context"
	"log"
	"os"

	"github.com/pkg/errors"
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

	if err := worker(ctx, client, hostname); err != nil {
		log.Fatalln(err)
	}
}

func worker(ctx context.Context, client *mq.Wrapper, hostname string) error {
	ch := client.Chan

	q, err := ch.QueueDeclare(
		"task_queue", // name
		false,        // durable
		false,        // delete when unused
		false,        // exclusive
		false,        // no-wait
		nil,          // arguments
	)
	if err != nil {
		return errors.Wrap(err, "failed to declare a queue")
	}

	if err := ch.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	); err != nil {
		return errors.Wrap(err, "failed to set QoS")
	}

	messages, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		return errors.Wrap(err, "failed to register a consumer")
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
