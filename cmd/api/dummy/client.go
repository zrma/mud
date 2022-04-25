package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/pborman/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"

	"mud/pb"
)

const (
	address = "localhost:4503"
)

func main() {
	if err := run(); err != nil {
		log.Fatalln(err)
	}
}

func run() error {
	conn, err := grpc.Dial(
		address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			// keepalive settings - https://github.com/grpc/grpc/blob/master/doc/keepalive.md
			Time:                10 * time.Second,
			Timeout:             5 * time.Second,
			PermitWithoutStream: true,
		}),
	)
	if err != nil {
		return fmt.Errorf("grpc connection failed: %v", err)
	}
	defer func() {
		_ = conn.Close()
	}()

	c := pb.NewMudClient(conn)

	name := uuid.New()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	directions := []pb.Direction{
		pb.Direction_NORTH,
		pb.Direction_SOUTH,
		pb.Direction_EAST,
		pb.Direction_WEST,
	}

	var step = 0
	for ctx.Err() == nil {
		step = (step + 1) % len(directions)
		if reply, err := c.Move(ctx, &pb.MoveRequest{
			Player:    name,
			Direction: directions[step],
		}); err != nil {
			return fmt.Errorf("move failed: %v %v", err, reply.GetErr())
		} else {
			fmt.Println(reply)
		}
		time.Sleep(time.Second * 3)
	}

	fmt.Println("done")
	return nil
}
