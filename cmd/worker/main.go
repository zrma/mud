package main

import (
	"context"
	"log"
	"os"

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

	worker(ctx, client, hostname)
}

func worker(ctx context.Context, client *mq.Wrapper, hostname string) {
	ch := client.Chan

	q, err := ch.QueueDeclare(
		"task_queue", // name
		false,        // durable
		false,        // delete when unused
		false,        // exclusive
		false,        // no-wait
		nil,          // arguments
	)
	failOnError(err, "Failed to declare a queue")

	err = ch.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	failOnError(err, "Failed to set QoS")

	messages, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")

	for ctx.Err() == nil {
		select {
		case <-ctx.Done():
			break
		case message := <-messages:
			msg := message.Body
			log.Printf("[#%s] Received a message: %s\n", hostname, msg)

			if err := message.Ack(false); err != nil {
				log.Println("Ack failed", err)
			} else {
				log.Println("Ack succeeded")
			}
		}
	}
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}
