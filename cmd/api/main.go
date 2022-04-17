package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/streadway/amqp"

	"mud/pkg/k8s"
	"mud/pkg/mq"
)

const defaultAddr = ":8080"

func main() {
	addr := defaultAddr
	if p := os.Getenv("PORT"); p != "" {
		addr = ":" + p
	}
	log.Printf("Server listening on port %s", addr)

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
		log.Fatal(err)
	}

	h := handler{
		hostname: hostname,
		client:   client,
	}

	http.HandleFunc("/", h.home)

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Server listening error: %+v", err)
	}
}

type handler struct {
	client   *mq.Wrapper
	hostname string
}

func (h handler) home(w http.ResponseWriter, r *http.Request) {
	msg := fmt.Sprintf("Host[%s] received from %s, %s", h.hostname, r.RemoteAddr, r.Method)
	if err := createTask(h.client, msg); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		msg := "Error creating request: " + err.Error()
		w.WriteHeader(http.StatusOK)
		if n, err := w.Write([]byte(msg)); err != nil {
			log.Printf("Error writing response: %+v, %d", err, n)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	n, err := w.Write([]byte("API server " + h.hostname + " received request"))
	if err != nil {
		log.Printf("api failed to write response: %+v, %d", err, n)
	}
}

func createTask(client *mq.Wrapper, msg string) error {
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
		return fmt.Errorf("failed to declare a queue: %+v", err)
	}

	body := fmt.Sprintf("message: %s", msg)
	if err := ch.Publish(
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "text/plain",
			Body:         []byte(body),
		},
	); err != nil {
		return fmt.Errorf("failed to publish a message: %+v", err)
	}
	return nil
}
