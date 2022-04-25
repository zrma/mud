package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"

	"mud/pb"
	"mud/pkg/k8s"
	"mud/pkg/mq"
	"mud/pkg/service"
)

const defaultAddr = ":8080"

func main() {
	addr := defaultAddr
	if p := os.Getenv("PORT"); p != "" {
		addr = ":" + p
	}
	log.Printf("Server listening on port %s", addr)

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

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

	channel := client.Chan
	queue, err := channel.QueueDeclare(
		"task_queue", // name
		false,        // durable
		false,        // delete when unused
		false,        // exclusive
		false,        // no-wait
		nil,          // arguments
	)
	if err != nil {
		log.Fatal(fmt.Errorf("failed to declare a queue: %+v", err))
	}

	svc := service.Service{
		Channel: channel,
		Queue:   queue,
	}

	var opts []grpc.ServerOption
	opts = append(opts,
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle:     30 * time.Second,
			MaxConnectionAge:      1 * time.Minute,
			MaxConnectionAgeGrace: 5 * time.Second,
			// pings the client to see if the transport is still alive.
			Time:    20 * time.Second,
			Timeout: 5 * time.Second,
		}),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             12 * time.Second,
			PermitWithoutStream: true,
		}),
	)
	svr := grpc.NewServer(opts...)
	pb.RegisterMudServer(svr, &svc)

	if err := svr.Serve(listener); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
