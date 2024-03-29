package main

import (
	"log"
	"net"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"

	"github.com/zrma/mud/pb"
	"github.com/zrma/mud/pkg/k8s"
	"github.com/zrma/mud/pkg/mq"
	"github.com/zrma/mud/pkg/service"
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

	publisher, err := client.Publisher("task_queue")
	if err != nil {
		log.Fatalln(err)
	}

	svc := service.Service{
		Publisher: publisher,
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
