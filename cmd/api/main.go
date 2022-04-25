package main

import (
	"context"
	"fmt"
	"log"
	"mime"
	"net"
	"os"
	"time"

	"github.com/streadway/amqp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/protobuf/encoding/protojson"

	"mud/pb"
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

	service := service{
		channel: channel,
		queue:   queue,
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
	server := grpc.NewServer(opts...)
	pb.RegisterMudServer(server, &service)

	if err := server.Serve(listener); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

type service struct {
	pb.UnimplementedMudServer

	queue   amqp.Queue
	channel *amqp.Channel
}

func (s service) Move(_ context.Context, req *pb.MoveRequest) (*pb.MoveReply, error) {
	if err := s.createTask(req); err != nil {
		return &pb.MoveReply{
			Player: req.GetPlayer(),
			Ok:     false,
			Err:    err.Error(),
		}, nil
	}

	return &pb.MoveReply{
		Player: req.GetPlayer(),
		Ok:     true,
	}, nil
}

func (s service) createTask(req *pb.MoveRequest) error {
	body, err := protojson.Marshal(req)
	if err != nil {
		return err
	}

	if err := s.channel.Publish(
		"",           // exchange
		s.queue.Name, // routing key
		false,        // mandatory
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  mime.TypeByExtension(".txt"),
			Body:         body,
		},
	); err != nil {
		return fmt.Errorf("failed to publish a message: %+v", err)
	}
	return nil
}
