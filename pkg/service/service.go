package service

import (
	"context"
	"fmt"
	"mime"

	"github.com/streadway/amqp"
	"google.golang.org/protobuf/encoding/protojson"

	"mud/pb"
)

type Service struct {
	pb.UnimplementedMudServer

	Queue   amqp.Queue
	Channel *amqp.Channel
}

func (s Service) Move(_ context.Context, req *pb.MoveRequest) (*pb.MoveReply, error) {
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

func (s Service) createTask(req *pb.MoveRequest) error {
	body, err := protojson.Marshal(req)
	if err != nil {
		return err
	}

	if err := s.Channel.Publish(
		"",           // exchange
		s.Queue.Name, // routing key
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
